package interactive

import (
	"context"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
)

type model struct {
	plan                *stellarstation.Plan
	operationStart      time.Time
	operationStop       time.Time
	withinOperationTime bool

	txStateTable      *table.Table
	txStateView       TransmitterState
	rxStateTable      *table.Table
	rxStateView       ReceiverState
	dataStateTable    *table.Table
	dataStateView     DataState
	antennaStateTable *table.Table
	antennaStateView  AntennaState

	streamID          string
	streamClosed      bool
	streamError       error
	lastCommandSentAt time.Time

	help       help.Model
	helpKeyMap helpKeyMap

	debugMode     bool
	debugLog      string
	teaUpdates    uint64
	debugViewport viewport.Model
}

var tableWidth = 120

func initialModel(
	ctx context.Context,
	client stellarstation.StellarStationServiceClient,
	plan *stellarstation.Plan,
	debugMode bool,
) model {
	txStateTable := table.New().Width(tableWidth).Headers(
		"Sweep", "Modulation", "Carrier", "Idle Pattern",
	).Border(lipgloss.RoundedBorder()).StyleFunc(tableStyleFunc)

	rxStateTable := table.New().Width(tableWidth).Headers(
		"Phase/Bit Lock", "Frame Lock", "Normalized SNR",
	).Border(lipgloss.RoundedBorder()).StyleFunc(tableStyleFunc)

	dataStateTable := table.New().Width(tableWidth).Headers(
		"Total Payload Bytes",
	).Border(lipgloss.RoundedBorder()).StyleFunc(tableStyleFunc)

	antennaStateTable := table.New().Width(tableWidth).Headers(
		"Est. Max Elevation", "Elevation", "Azimuth",
	).Border(lipgloss.RoundedBorder()).StyleFunc(tableStyleFunc)

	debugvp := viewport.New(tableWidth, 7)
	debugvp.Style = viewportStyleBlue

	opStart := plan.GetAosTime().AsTime()
	opStop := plan.GetLosTime().AsTime()

	helpModel := help.New()
	helpModel.Styles.FullDesc = helpStyleDescription
	helpModel.Styles.ShortDesc = helpStyleDescription

	helpModel.Styles.FullKey = helpStyleKey
	helpModel.Styles.ShortKey = helpStyleKey

	helpModel.Styles.FullSeparator = helpStyleSeparator
	helpModel.Styles.ShortSeparator = helpStyleSeparator

	helpModel.ShowAll = true
	// helpModel.Width = tableWidth

	inboundCrc32cTable := crc32.MakeTable(crc32.Castagnoli)
	inboundCrc32cHash := crc32.New(inboundCrc32cTable)

	// outboundCrc32cTable := crc32.MakeTable(crc32.Castagnoli)
	// outboundCrc32cHash := crc32.New(outboundCrc32cTable)

	stateMux.Lock()
	streamState.inboundCrc32c = inboundCrc32cHash
	// streamState.outboundCrc32c = outboundCrc32cHash
	stateMux.Unlock()

	go startStream(ctx, plan, client)

	return model{
		operationStart: opStart,
		operationStop:  opStop,

		plan:              plan,
		txStateTable:      txStateTable,
		rxStateTable:      rxStateTable,
		dataStateTable:    dataStateTable,
		antennaStateTable: antennaStateTable,
		antennaStateView: AntennaState{
			estMaxElevation: plan.GetMaxElevationDegrees(),
		},

		help:       helpModel,
		helpKeyMap: defaultKeyMap(),

		debugMode:     debugMode,
		debugViewport: debugvp,
	}
}

func (m model) Init() tea.Cmd {
	return startScreenInterval(time.Second)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var subCmds []tea.Cmd

	m.withinOperationTime = time.Now().After(m.operationStart.Add(-5*time.Second)) && time.Now().Before(m.operationStop.Add(5*time.Second))

	// synchronize the stream state with the tui view state
	stateMux.Lock()
	m.txStateView.sweepOn = streamState.currentTxState.sweepOn
	m.txStateView.modulationOn = streamState.currentTxState.modulationOn
	m.txStateView.carrierOn = streamState.currentTxState.carrierOn
	m.txStateView.idlePatternOn = streamState.currentTxState.idlePatternOn

	m.rxStateView.phaseBitLock = streamState.currentRxState.phaseBitLock
	m.rxStateView.frameLock = streamState.currentRxState.frameLock
	m.rxStateView.normalizedSnr = streamState.currentRxState.normalizedSnr

	m.dataStateView.totalPayloadBytes = streamState.totalPayloadBytes

	m.antennaStateView.elevation = streamState.currentAntennaState.elevation
	m.antennaStateView.azimuth = streamState.currentAntennaState.azimuth

	m.streamID = streamState.streamId
	m.streamClosed = streamState.closed
	m.dataStateView.receivedEndTelemetryMessage = streamState.receivedEndTelemetryMessage

	m.dataStateView.inboundCrc32cStr = fmt.Sprintf("%v", streamState.inboundCrc32c.Sum32())
	// m.dataStateView.outboundCrc32cStr = fmt.Sprintf("%v", streamState.outboundCrc32c.Sum32())

	if streamState.err != nil {
		m.debugLog = prependLine(m.debugLog, fmt.Sprintf("stream err: %v", streamState.err.Error()))
		m.streamError = streamState.err
		streamState.err = nil
	}
	stateMux.Unlock()

	// the TUI framework does not have a concept of key up
	// This can cause key repeats and such to trigger double presses of commands
	// or if the user JUST sent something, we'll block it
	recentlySentCmd := !m.lastCommandSentAt.IsZero() && time.Since(m.lastCommandSentAt) <= 100*time.Millisecond

	m.helpKeyMap = m.helpKeyMap.Update(m)

	m.teaUpdates += 1
	switch msg := msg.(type) {
	case configurationChangeSent:
		b := string(msg)
		if len(b) == 0 {
			m.debugLog = prependLine(m.debugLog, "some configuration change sent")
		} else {
			m.debugLog = prependLine(m.debugLog, b)
		}
		m.lastCommandSentAt = time.Now()
		m.dataStateView.configurationChangeSentCount += 1
	case commandSent:
		b := string(msg)
		if len(b) == 0 {
			m.debugLog = prependLine(m.debugLog, "some command sent")
		} else {
			m.debugLog = prependLine(m.debugLog, b)
		}
		m.lastCommandSentAt = time.Now()
		m.dataStateView.commandSentCount += 1
	case errMsg:
		if msg.err != nil {
			m.debugLog = prependLine(m.debugLog, fmt.Sprintf("error: %s", msg.err.Error()))
		}
	case screenIntervalTick:
		// re-emit same tick duration
		return m, startScreenInterval(msg.duration)
	case tea.KeyMsg:

		switch {
		case key.Matches(msg, m.helpKeyMap.Quit):
			stateMux.Lock()
			if STREAM_CLIENT != nil {
				_ = STREAM_CLIENT.CloseSend()
			}
			stateMux.Unlock()
			return m, tea.Quit
		}

		if recentlySentCmd {
			m.debugLog = prependLine(m.debugLog, "key input skipped, recently sent command")
			break
		}

		if m.streamClosed || STREAM_CLIENT == nil {
			m.debugLog = prependLine(m.debugLog, "key input skipped, stream not open")
			break
		}

		switch {
		case key.Matches(msg, m.helpKeyMap.SweepEnable):
			cmd := func() tea.Msg {
				return sweep(true, m)
			}
			subCmds = append(subCmds, cmd)
		case key.Matches(msg, m.helpKeyMap.SweepDisable):
			cmd := func() tea.Msg {
				return sweep(false, m)
			}
			subCmds = append(subCmds, cmd)
		case key.Matches(msg, m.helpKeyMap.ModulationEnable):
			cmd := func() tea.Msg {
				return modulation(true, m)
			}
			subCmds = append(subCmds, cmd)
		case key.Matches(msg, m.helpKeyMap.ModulationDisable):
			cmd := func() tea.Msg {
				return modulation(false, m)
			}
			subCmds = append(subCmds, cmd)
		case key.Matches(msg, m.helpKeyMap.CarrierEnable):
			cmd := func() tea.Msg {
				return carrier(true, m)
			}
			subCmds = append(subCmds, cmd)
		case key.Matches(msg, m.helpKeyMap.CarrierDisable):
			cmd := func() tea.Msg {
				return carrier(false, m)
			}
			subCmds = append(subCmds, cmd)
		case key.Matches(msg, m.helpKeyMap.IdlePatternEnable):
			cmd := func() tea.Msg {
				return idlePattern(true, m)
			}
			subCmds = append(subCmds, cmd)
		case key.Matches(msg, m.helpKeyMap.IdlePatternDisable):
			cmd := func() tea.Msg {
				return idlePattern(false, m)
			}
			subCmds = append(subCmds, cmd)
		default:
			m.debugLog = prependLine(m.debugLog, fmt.Sprintf("unexpected key: '%s'", msg.String()))
		}
	}

	if m.debugMode {
		var vpCmd tea.Cmd

		m.debugViewport.SetContent(m.debugLog)
		m.debugViewport, vpCmd = m.debugViewport.Update(msg)
		subCmds = append(subCmds, vpCmd)
	}

	return m, tea.Batch(subCmds...)
}

func (m model) View() string {
	m.txStateTable.Data(
		table.NewStringData(m.txStateView.toStringColumns()),
	)

	m.rxStateTable.Data(
		table.NewStringData(m.rxStateView.toStringColumns()),
	)

	m.dataStateTable.Data(
		table.NewStringData(m.dataStateView.toStringColumns()),
	)

	m.antennaStateTable.Data(
		table.NewStringData(m.antennaStateView.toStringColumns()),
	)

	var builder strings.Builder
	parameters := make([]any, 0, 12)

	if !m.streamClosed && time.Now().Before(m.plan.LosTime.AsTime().Add(20*time.Second)) {
		builder.WriteString("%s: \n%v\n")
		parameters = append(parameters, boldStyle.Render("Transmitter status"), m.txStateTable.Render())

		builder.WriteString("%s: \n%v\n")
		parameters = append(parameters, boldStyle.Render("Receiver Status"), m.rxStateTable.Render())

		builder.WriteString("%s: \n%v\n")
		parameters = append(parameters, boldStyle.Render("Antenna"), m.antennaStateTable.Render())
	}

	builder.WriteString("%s: \n%v\n")
	parameters = append(parameters, boldStyle.Render("Data"), m.dataStateTable.Render())

	builder.WriteString("%50s: %v\n")
	parameters = append(parameters, boldStyle.Render("Plan ID"), m.plan.GetId())
	builder.WriteString("%50s: %v\n")
	parameters = append(parameters, boldStyle.Render("Stream ID"), m.streamID)
	builder.WriteString("%50s: %s (%s)\n")
	parameters = append(parameters,
		boldStyle.Render("Operation Start Time"),
		textDimmer(!m.withinOperationTime, m.operationStart.Format(time.RFC3339)),
		textDimmer(!m.withinOperationTime, time.Since(m.operationStart).String()),
	)
	builder.WriteString("%50s: %s (%s)\n")
	parameters = append(parameters,
		boldStyle.Render("Operation Stop Time"),
		textDimmer(!m.withinOperationTime, m.operationStop.Format(time.RFC3339)),
		textDimmer(!m.withinOperationTime, time.Since(m.operationStop).String()),
	)

	builder.WriteString("%50s: %v\n")
	parameters = append(parameters,
		boldStyle.Render("Configuration Changes Sent"),
		m.dataStateView.configurationChangeSentCount,
	)

	if m.streamClosed ||
		m.dataStateView.receivedEndTelemetryMessage ||
		time.Now().After(m.plan.LosTime.AsTime().Add(20*time.Second)) ||
		m.streamClosed {

		builder.WriteString("%50s: %v\n")
		parameters = append(parameters, boldStyle.Render("Inbound Data CRC32C"), m.dataStateView.inboundCrc32cStr)

		// builder.WriteString("%30s: %v\n")
		// parameters = append(parameters, boldStyle.Render("Outbound Data CRC32C"), m.dataStateView.outboundCrc32cStr)
	}

	if m.debugMode {
		builder.WriteString("%50s: %v\n")
		parameters = append(parameters, boldStyle.Render("Screen Updates"), m.teaUpdates)
		builder.WriteString("%s: \n%v\n")
		parameters = append(parameters, boldStyle.Render("Debug"), m.debugViewport.View())
	}

	builder.WriteString("\n")
	builder.WriteString(m.help.View(m.helpKeyMap))

	builder.WriteString("\n")

	return fmt.Sprintf(
		builder.String(),
		parameters...,
	)

}
