package interactive

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type TransmitterState struct {
	sweepOn       bool
	modulationOn  bool
	carrierOn     bool
	idlePatternOn bool
}

func (cs TransmitterState) toStringColumns() []string {
	return []string{
		strconv.FormatBool(cs.sweepOn),
		strconv.FormatBool(cs.modulationOn),
		strconv.FormatBool(cs.carrierOn),
		strconv.FormatBool(cs.idlePatternOn),
	}
}

type ReceiverState struct {
	phaseBitLock  bool
	frameLock     bool
	normalizedSnr float64
}

func (cs ReceiverState) toStringColumns() []string {
	return []string{
		strconv.FormatBool(cs.phaseBitLock),
		strconv.FormatBool(cs.frameLock),
		strconv.FormatFloat(cs.normalizedSnr, 'f', 4, 64),
	}
}

type model struct {
	viewport          viewport.Model
	ssc               stellarstation.StellarStationServiceClient
	stream            stellarstation.StellarStationService_OpenSatelliteStreamClient
	plan              *stellarstation.Plan
	txStateTable      *table.Table
	rxStateTable      *table.Table
	txStateView       TransmitterState
	rxStateView       ReceiverState
	streamId          string
	totalPayloadBytes uint64
	streamClosed      bool
	streamError       error
	debug             string
	updates           uint64
	lastCommandSentAt time.Time
}

type StreamState struct {
	currentTxState    TransmitterState
	currentRxState    ReceiverState
	streamId          string
	totalPayloadBytes uint64
	closed            bool
	err               error
}

var stateMux sync.Mutex
var streamState StreamState

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

type commandSent string

func idlePattern(enable bool, planId, satelliteId, groundStationId, streamId string, s stellarstation.StellarStationService_OpenSatelliteStreamClient) tea.Msg {
	stateMux.Lock()
	err := s.Send(&stellarstation.SatelliteStreamRequest{
		SatelliteId:     satelliteId,
		PlanId:          planId,
		GroundStationId: groundStationId,
		StreamId:        streamId,
		Request: &stellarstation.SatelliteStreamRequest_GroundStationConfigurationRequest{
			GroundStationConfigurationRequest: &stellarstation.GroundStationConfigurationRequest{
				TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
					EnableIdlePattern: &wrapperspb.BoolValue{
						Value: enable,
					},
				},
			},
		},
	})
	defer stateMux.Unlock()

	if err != nil {
		return errMsg{err}
	}

	return commandSent(fmt.Sprintf("idle %v", enable))
}

func modulation(enable bool, planId, satelliteId, groundStationId, streamId string, s stellarstation.StellarStationService_OpenSatelliteStreamClient) tea.Msg {
	stateMux.Lock()
	err := s.Send(&stellarstation.SatelliteStreamRequest{
		SatelliteId:     satelliteId,
		PlanId:          planId,
		GroundStationId: groundStationId,
		StreamId:        streamId,
		Request: &stellarstation.SatelliteStreamRequest_GroundStationConfigurationRequest{
			GroundStationConfigurationRequest: &stellarstation.GroundStationConfigurationRequest{
				TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
					EnableIfModulation: &wrapperspb.BoolValue{
						Value: enable,
					},
				},
			},
		},
	})
	defer stateMux.Unlock()

	if err != nil {
		return errMsg{err}
	}

	return commandSent(fmt.Sprintf("modulation %v", enable))
}

func carrier(enable bool, planId, satelliteId, groundStationId, streamId string, s stellarstation.StellarStationService_OpenSatelliteStreamClient) tea.Msg {
	stateMux.Lock()
	err := s.Send(&stellarstation.SatelliteStreamRequest{
		SatelliteId:     satelliteId,
		PlanId:          planId,
		GroundStationId: groundStationId,
		StreamId:        streamId,
		Request: &stellarstation.SatelliteStreamRequest_GroundStationConfigurationRequest{
			GroundStationConfigurationRequest: &stellarstation.GroundStationConfigurationRequest{
				TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
					EnableCarrier: &wrapperspb.BoolValue{
						Value: enable,
					},
				},
			},
		},
	})
	defer stateMux.Unlock()

	if err != nil {
		return errMsg{err}
	}

	return commandSent(fmt.Sprintf("carrier %v", enable))
}

func sweep(enable bool, planId, satelliteId, groundStationId, streamId string, s stellarstation.StellarStationService_OpenSatelliteStreamClient) tea.Msg {
	stateMux.Lock()
	err := s.Send(&stellarstation.SatelliteStreamRequest{
		SatelliteId:     satelliteId,
		PlanId:          planId,
		GroundStationId: groundStationId,
		StreamId:        streamId,
		Request: &stellarstation.SatelliteStreamRequest_GroundStationConfigurationRequest{
			GroundStationConfigurationRequest: &stellarstation.GroundStationConfigurationRequest{
				TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
					EnableIfSweep: &wrapperspb.BoolValue{
						Value: enable,
					},
				},
			},
		},
	})
	defer stateMux.Unlock()

	if err != nil {
		return errMsg{err}
	}

	return commandSent(fmt.Sprintf("sweep %v", enable))
}

func initialModel(ctx context.Context, client stellarstation.StellarStationServiceClient, plan *stellarstation.Plan) model {
	vp := viewport.New(80, 6)
	vp.SetContent(`
		press 1 to enable sweep, press 2 to disable sweep
		press 3 to enable modulation, press 4 to disable modulation
		press 5 to enable carrier, press 6 to disable carrier
		press 7 to enable idle pattern, press 8 to disable idle pattern
	`)
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Copy().Foreground(lipgloss.Color("252")).Bold(true)

	txStateTable := table.New().Headers(
		"Sweep", "Modulation", "Carrier", "Idle Pattern",
	).Border(lipgloss.NormalBorder()).StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}

		return baseStyle.Copy().Foreground(lipgloss.Color("252"))
	})

	rxStateTable := table.New().Headers(
		"Phase/Bit Lock", "Frame Lock", "Normalized SNR",
	).Border(lipgloss.NormalBorder()).StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}

		return baseStyle.Copy().Foreground(lipgloss.Color("252"))
	})

	stream, err := client.OpenSatelliteStream(ctx)
	if err != nil {
		log.Fatalf("problem connecting to stream: %v", err)
	}

	satelliteStreamRequest := stellarstation.SatelliteStreamRequest{
		SatelliteId:              plan.GetSatelliteId(),
		StreamId:                 "",
		ResumeStreamMessageAckId: "",
		EnableFlowControl:        true,
		PlanId:                   plan.GetId(),
		GroundStationId:          plan.GetGroundStationId(),
	}

	err = stream.Send(&satelliteStreamRequest)
	if err != nil {
		log.Fatalf("problem requesting stream: %v", err)
	}
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				stateMux.Lock()
				streamState.closed = true
				streamState.err = err
				stateMux.Unlock()
				return
			}
			stateMux.Lock()
			if streamState.streamId == "" {
				streamState.streamId = msg.GetStreamId()
			}

			switch msg.GetResponse().(type) {
			case *stellarstation.SatelliteStreamResponse_ReceiveTelemetryResponse:
				telemetryResponse := msg.GetReceiveTelemetryResponse()
				if telemetryResponse == nil {
					stateMux.Unlock()
					continue
				}
				for _, telemetry := range telemetryResponse.Telemetry {
					if telemetry == nil {
						break
					}
					streamState.totalPayloadBytes += uint64(len(telemetry.Data))
				}

				ackID := telemetryResponse.GetMessageAckId()
				ackRequest := stellarstation.SatelliteStreamRequest{
					SatelliteId: plan.GetSatelliteId(),
					Request: &stellarstation.SatelliteStreamRequest_TelemetryReceivedAck{
						TelemetryReceivedAck: &stellarstation.ReceiveTelemetryAck{
							MessageAckId:      ackID,
							ReceivedTimestamp: timestamppb.Now(),
						},
					},
				}
				_ = stream.Send(&ackRequest)
			case *stellarstation.SatelliteStreamResponse_StreamEvent:
				monitoring := msg.GetStreamEvent().GetPlanMonitoringEvent().GetGroundStationState()
				if monitoring == nil {
					stateMux.Unlock()
					continue
				}
				if txState := monitoring.GetTransmitter(); txState != nil {
					streamState.currentTxState.carrierOn = txState.GetIsCarrierEnabled().GetValue()
					streamState.currentTxState.idlePatternOn = txState.GetIsIdlePatternEnabled().GetValue()
					streamState.currentTxState.sweepOn = txState.GetIsIfSweepEnabled().GetValue()
					streamState.currentTxState.modulationOn = txState.GetIsModulationEnabled().GetValue()
				}
				if rxState := monitoring.GetReceiver(); rxState != nil {
					streamState.currentRxState.phaseBitLock = rxState.GetIsBitSynchronizerLocked() || rxState.GetIsPhaseLocked()
					streamState.currentRxState.frameLock = rxState.GetIsFrameSynchronizerLocked()
					streamState.currentRxState.normalizedSnr = rxState.GetNormalizedSnr()
				}
			}
			stateMux.Unlock()
		}
	}()

	return model{
		ssc:          client,
		stream:       stream,
		plan:         plan,
		viewport:     vp,
		txStateTable: txStateTable,
		rxStateTable: rxStateTable,
	}
}

type (
	tickMsg struct{}
)

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(msg)

	stateMux.Lock()
	m.txStateView = TransmitterState{
		sweepOn:       streamState.currentTxState.sweepOn,
		modulationOn:  streamState.currentTxState.modulationOn,
		carrierOn:     streamState.currentTxState.carrierOn,
		idlePatternOn: streamState.currentTxState.idlePatternOn,
	}

	m.rxStateView = ReceiverState{
		phaseBitLock:  streamState.currentRxState.phaseBitLock,
		frameLock:     streamState.currentRxState.frameLock,
		normalizedSnr: streamState.currentRxState.normalizedSnr,
	}
	m.streamId = streamState.streamId
	m.totalPayloadBytes = streamState.totalPayloadBytes
	m.streamClosed = streamState.closed
	m.streamError = streamState.err
	stateMux.Unlock()

	m.updates += 1
	switch msg := msg.(type) {
	case commandSent:
		b := string(msg)
		if len(b) == 0 {
			m.debug = "unknown command sent"
		} else {
			m.debug = b
		}
	case errMsg:
		if msg.err != nil {
			m.streamError = msg.err
		}
	case tickMsg:
		return m, tick()
	case tea.KeyMsg:

		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c":
			if m.stream != nil {
				_ = m.stream.CloseSend()
			}
			return m, tea.Quit
		case "1", "2":
			if m.stream != nil && (m.lastCommandSentAt.IsZero() || time.Since(m.lastCommandSentAt) > time.Second) {
				m.debug = "sending sweep"
				enable := msg.String() == "1"
				cmd := func() tea.Msg {
					return sweep(
						enable,
						m.plan.Id,
						m.plan.SatelliteId,
						m.plan.GroundStationId,
						m.streamId,
						m.stream,
					)
				}
				m.lastCommandSentAt = time.Now()

				return m, cmd
			} else {
				m.debug = "skip sweep"
			}
		case "3", "4":
			if m.stream != nil && (m.lastCommandSentAt.IsZero() || time.Since(m.lastCommandSentAt) > time.Second) {
				m.debug = "sending mod"
				enable := msg.String() == "3"
				cmd := func() tea.Msg {
					return modulation(
						enable,
						m.plan.Id,
						m.plan.SatelliteId,
						m.plan.GroundStationId,
						m.streamId,
						m.stream,
					)
				}
				m.lastCommandSentAt = time.Now()

				return m, cmd
			} else {
				m.debug = "skip mod"
			}

		case "5", "6":
			if m.stream != nil && (m.lastCommandSentAt.IsZero() || time.Since(m.lastCommandSentAt) > time.Second) {
				m.debug = "sending carrier"
				enable := msg.String() == "5"
				cmd := func() tea.Msg {
					return carrier(
						enable,
						m.plan.Id,
						m.plan.SatelliteId,
						m.plan.GroundStationId,
						m.streamId,
						m.stream,
					)
				}
				m.lastCommandSentAt = time.Now()

				return m, cmd
			} else {
				m.debug = "skip carrier"
			}

		case "7", "8":
			if m.stream != nil && (m.lastCommandSentAt.IsZero() || time.Since(m.lastCommandSentAt) > time.Second) {
				m.debug = "sending idle"
				enable := msg.String() == "7"
				cmd := func() tea.Msg {
					return idlePattern(
						enable,
						m.plan.Id,
						m.plan.SatelliteId,
						m.plan.GroundStationId,
						m.streamId,
						m.stream,
					)
				}
				m.lastCommandSentAt = time.Now()

				return m, cmd
			} else {
				m.debug = "skip idle"
			}
		default:
			m.debug = fmt.Sprintf("unhandled key: '%s'", msg.String())
		}
	}

	return m, tea.Batch(vpCmd)
}

func (m model) View() string {

	m.txStateTable.Data(
		table.NewStringData(m.txStateView.toStringColumns()),
	)

	m.rxStateTable.Data(
		table.NewStringData(m.rxStateView.toStringColumns()),
	)

	return fmt.Sprintf(
		"transmitter status:\n%s\nreceiver status:\n%s\n\n%s\nStream ID: %s\nTotal Bytes: %v\nAOS %v (%v)\nLOS %v (%v)\nstream done? %v\nstream err: %v\ndebug: %s\nlast cmd %s\nscreen updates: %v\n%s",
		m.txStateTable.Render(),
		m.rxStateTable.Render(),
		m.viewport.View(),
		m.streamId,
		m.totalPayloadBytes,
		m.plan.GetAosTime().AsTime(),
		time.Now().Sub(m.plan.GetAosTime().AsTime()),
		m.plan.GetLosTime().AsTime(),
		time.Now().Sub(m.plan.GetLosTime().AsTime()),
		strconv.FormatBool(m.streamClosed),
		m.streamError,
		m.debug,
		m.lastCommandSentAt.Format(time.RFC3339),
		m.updates,
		"Press ctrl+c to quit.",
	) + "\n\n"
}
