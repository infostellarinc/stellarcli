package interactive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func configurationChange(
	debugMsg string,
	req *stellarstation.GroundStationConfigurationRequest,
	m model,
) tea.Msg {
	stateMux.Lock()
	defer stateMux.Unlock()

	if STREAM_CLIENT == nil {
		return errMsg{err: errors.New("stream client not active")}
	}

	err := STREAM_CLIENT.Send(&stellarstation.SatelliteStreamRequest{
		SatelliteId:     m.plan.GetSatelliteId(),
		PlanId:          m.plan.GetId(),
		GroundStationId: m.plan.GetGroundStationId(),
		StreamId:        m.streamID,
		Request: &stellarstation.SatelliteStreamRequest_GroundStationConfigurationRequest{
			GroundStationConfigurationRequest: req,
		},
	})
	if err != nil {
		return errMsg{fmt.Errorf("could not send configuration change: %w", err)}
	}

	return configurationChangeSent(fmt.Sprintf("configuration change: %v", debugMsg))
}

func idlePattern(enable bool, m model) tea.Msg {
	debugMsg := "idle pattern "
	if enable {
		debugMsg += "enable"
	} else {
		debugMsg += "disable"
	}

	return configurationChange(debugMsg, &stellarstation.GroundStationConfigurationRequest{
		TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
			EnableIdlePattern: &wrapperspb.BoolValue{
				Value: enable,
			},
		},
	}, m)
}

func modulation(enable bool, m model) tea.Msg {
	debugMsg := "modulation "
	if enable {
		debugMsg += "enable"
	} else {
		debugMsg += "disable"
	}

	return configurationChange(debugMsg, &stellarstation.GroundStationConfigurationRequest{
		TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
			EnableIfModulation: &wrapperspb.BoolValue{
				Value: enable,
			},
		},
	}, m)
}

func carrier(enable bool, m model) tea.Msg {
	debugMsg := "carrier "
	if enable {
		debugMsg += "enable"
	} else {
		debugMsg += "disable"
	}

	return configurationChange(debugMsg, &stellarstation.GroundStationConfigurationRequest{
		TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
			EnableCarrier: &wrapperspb.BoolValue{
				Value: enable,
			},
		},
	}, m)
}

func sweep(enable bool, m model) tea.Msg {
	debugMsg := "sweep "
	if enable {
		debugMsg += "enable"
	} else {
		debugMsg += "disable"
	}

	return configurationChange(debugMsg, &stellarstation.GroundStationConfigurationRequest{
		TransmitterConfigurationRequest: &stellarstation.TransmitterConfigurationRequest{
			EnableIfSweep: &wrapperspb.BoolValue{
				Value: enable,
			},
		},
	}, m)
}

func streamHandler(plan *stellarstation.Plan) error {
	for {
		msg, err := STREAM_CLIENT.Recv()
		if err != nil {
			stateMux.Lock()
			streamState.closed = true
			streamState.err = err
			stateMux.Unlock()
			return err
		}
		stateMux.Lock()
		if streamState.streamId == "" || streamState.streamId != msg.GetStreamId() {
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
				_, _ = streamState.inboundCrc32c.Write(telemetry.Data)
				streamState.totalPayloadBytes += uint64(len(telemetry.Data))
			}
			if len(telemetryResponse.GetTelemetry()) == 1 && len(telemetryResponse.GetTelemetry()[0].GetData()) == 0 {
				streamState.receivedEndTelemetryMessage = true
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
			streamState.lastAck = ackID
			streamState.err = STREAM_CLIENT.Send(&ackRequest)
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

			if antennaState := monitoring.GetAntenna(); antennaState != nil {
				streamState.currentAntennaState.elevation = antennaState.Elevation.Measured
				streamState.currentAntennaState.azimuth = antennaState.Azimuth.Measured
			}
		}
		stateMux.Unlock()
	}
}

func startStream(ctx context.Context, plan *stellarstation.Plan, client stellarstation.StellarStationServiceClient) {
	defer func() {
		stateMux.Lock()
		streamState.closed = true
		streamState.err = fmt.Errorf("stream closed")
		stateMux.Unlock()
	}()
	failureCount := 0
	for failureCount < 20 {
		stateMux.Lock()
		stream, err := client.OpenSatelliteStream(ctx)
		if err != nil {
			failureCount += 1
			streamState.err = err
			stateMux.Unlock()
			time.Sleep(500 * time.Second)
			continue
		}

		STREAM_CLIENT = stream

		satelliteStreamRequest := stellarstation.SatelliteStreamRequest{
			SatelliteId:              plan.GetSatelliteId(),
			StreamId:                 streamState.streamId,
			ResumeStreamMessageAckId: streamState.lastAck,
			EnableFlowControl:        true,
			PlanId:                   plan.GetId(),
			GroundStationId:          plan.GetGroundStationId(),
		}

		if err = STREAM_CLIENT.Send(&satelliteStreamRequest); err != nil {
			streamState.err = err
			failureCount += 1
			stateMux.Unlock()
			time.Sleep(500 * time.Second)
			continue
		} else {
			streamState.closed = false
		}

		stateMux.Unlock()

		if err := streamHandler(plan); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			} else if errors.Is(err, io.EOF) {
				failureCount += 1
				time.Sleep(500 * time.Second)
			} else if err != nil {
				streamState.err = err
				failureCount += 1
				time.Sleep(500 * time.Second)
			}

		}
	}
}
