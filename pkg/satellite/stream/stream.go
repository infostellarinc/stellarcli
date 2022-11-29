// Copyright © 2018 Infostellar, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stream

import (
	"bufio"
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/pkg/util/collection"
)

const (
	OPEN   uint32 = 0
	CLOSED uint32 = 1
)

const MaxElapsedTime = 60 * time.Second

var metrics MetricsCollector

type SatelliteStreamOptions struct {
	AcceptedFraming []stellarstation.Framing
	SatelliteID     string
	StreamId        string
	PlanId          string
	GroundStationId string
	IsDebug         bool
	IsVerbose       bool
	ShowStats       bool
	TelemetryFile   *os.File

	CorrectOrder   bool
	DelayThreshold time.Duration

	EnableAutoClose bool
}

type SatelliteStream interface {
	Send(payload []byte) error

	io.Closer
}

type satelliteStream struct {
	acceptedFraming []stellarstation.Framing

	satelliteId     string
	stream          stellarstation.StellarStationService_OpenSatelliteStreamClient
	conn            *grpc.ClientConn
	streamId        string
	planId          string
	groundStationId string

	receiveChan           chan<- []byte
	receiveLoopClosedChan chan struct{}

	telemetryFileWriter *bufio.Writer

	state         uint32
	isDebug       bool
	isVerbose     bool
	showStats     bool
	telemetryFile *os.File

	correctOrder   bool
	delayThreshold time.Duration
	flushTimer     *time.Timer
	mu             sync.Mutex

	enableAutoClose bool
}

// OpenSatelliteStream opens a stream to a satellite over the StellarStation API.
func OpenSatelliteStream(o *SatelliteStreamOptions, receiveChan chan<- []byte) (SatelliteStream, func(), error) {
	satelliteStream := &satelliteStream{
		acceptedFraming:       o.AcceptedFraming,
		satelliteId:           o.SatelliteID,
		streamId:              o.StreamId,
		planId:                o.PlanId,
		groundStationId:       o.GroundStationId,
		receiveChan:           receiveChan,
		state:                 OPEN,
		receiveLoopClosedChan: make(chan struct{}),
		isDebug:               o.IsDebug,
		isVerbose:             o.IsVerbose,
		showStats:             o.ShowStats,
		telemetryFile:         o.TelemetryFile,

		correctOrder:   o.CorrectOrder,
		delayThreshold: o.DelayThreshold,

		enableAutoClose: o.EnableAutoClose,
	}

	cleanup, err := satelliteStream.start()

	return satelliteStream, cleanup, err
}

// Send sends a packet to the satellite.
func (ss *satelliteStream) Send(payload []byte) error {
	satelliteStreamRequest := stellarstation.SatelliteStreamRequest{
		SatelliteId: ss.satelliteId,
		Request: &stellarstation.SatelliteStreamRequest_SendSatelliteCommandsRequest{
			SendSatelliteCommandsRequest: &stellarstation.SendSatelliteCommandsRequest{
				Command: [][]byte{payload},
			},
		},
	}

	log.Verbose("sent data: size: %d bytes\n", len(payload))

	return ss.stream.Send(&satelliteStreamRequest)
}

// Close closes the stream.
func (ss *satelliteStream) Close() error {
	atomic.StoreUint32(&ss.state, CLOSED)

	ss.stream.CloseSend()
	ss.conn.Close()

	<-ss.receiveLoopClosedChan

	ss.CloseFileWriter()

	return nil
}

func (ss *satelliteStream) CloseFileWriter() error {
	if ss.telemetryFileWriter != nil {
		ss.telemetryFileWriter.Flush()
		ss.telemetryFileWriter = nil
	}
	if ss.telemetryFile != nil {
		err := ss.telemetryFile.Close()
		ss.telemetryFile = nil
		return err
	}
	return nil
}

// send telemetryMessageAckId to support enableFlowControl feature
func (ss *satelliteStream) ackReceivedTelemetry(telemetryMessageAckId string) {
	if telemetryMessageAckId != "" {
		satelliteStreamRequest := stellarstation.SatelliteStreamRequest{
			SatelliteId: ss.satelliteId,
			Request: &stellarstation.SatelliteStreamRequest_TelemetryReceivedAck{
				TelemetryReceivedAck: &stellarstation.ReceiveTelemetryAck{
					MessageAckId:      telemetryMessageAckId,
					ReceivedTimestamp: timestampNow(),
				},
			},
		}
		log.Debug("sending ack index: %v", telemetryMessageAckId)
		ss.stream.Send(&satelliteStreamRequest)
	}
}

func (ss *satelliteStream) performAutoClose() {
	log.Printf("Stream auto-close conditions met - exiting")
	if ss.showStats {
		metrics.logReport()
	}
	ss.Close()
	os.Exit(0)
}

func (ss *satelliteStream) receiveLoop() {
	streamEndDetected := false

	if ss.enableAutoClose {
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for {
				<-ticker.C
				if streamEndDetected {
					ss.performAutoClose()
				}
			}
		}()
	}

	// Initialize exponential back off settings.
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = MaxElapsedTime

	// file writer for telemetry data
	if ss.telemetryFile != nil {
		ss.telemetryFileWriter = bufio.NewWriter(ss.telemetryFile)
	}
	telemetryMessageAckId := ""

	pq := collection.NewPriorityQueue((*stellarstation.Telemetry)(nil), func(i, j interface{}) bool {
		telemetry1 := i.(*stellarstation.Telemetry)
		telemetry2 := j.(*stellarstation.Telemetry)

		time1, err := ptypes.Timestamp(telemetry1.TimeFirstByteReceived)
		if err != nil {
			log.Fatal(err)
		}
		time2, err := ptypes.Timestamp(telemetry2.TimeFirstByteReceived)
		if err != nil {
			log.Fatal(err)
		}

		return !time1.After(time2)
	})

	ss.flushTimer = time.AfterFunc(ss.delayThreshold, func() {
		ss.mu.Lock()
		defer ss.mu.Unlock()

		// Flush data in the PriorityQueue half of the incoming data.
		numFlush := (pq.Len() + 1) / 2
		if numFlush > 0 {
			log.Printf("%d data in the queue, flushed %d.\n", pq.Len(), numFlush)
		}

		// Flush half of the data in the priority queue
		for i := 0; i < numFlush; i++ {
			telemetry := pq.Pop().(*stellarstation.Telemetry)
			ss.receiveChan <- telemetry.Data
			if ss.telemetryFileWriter != nil {
				if _, err := ss.telemetryFileWriter.Write(telemetry.Data); err != nil {
					panic(err)
				}
			}
		}
		ss.flushTimer.Reset(ss.delayThreshold)
	})

	for {
		streamResponse, err := ss.stream.Recv()
		if atomic.LoadUint32(&ss.state) == CLOSED {
			// Closed, so just shutdown the loop.
			close(ss.receiveLoopClosedChan)
			return
		}
		if err != nil {
			log.Println(err)
			log.Println("reconnecting to the API stream.")

			reconnectError := backoff.RetryNotify(func() error {
				err := ss.openStream(telemetryMessageAckId)
				if err != nil {
					return err
				}

				response, err := ss.stream.Recv()
				if err != nil {
					return err
				}
				streamResponse = response

				return nil
			}, backOff,
				func(e error, duration time.Duration) {
					log.Printf("%s. Automatically retrying in %v", e, duration)
				})
			if reconnectError != nil {
				// This explicit cleanup is not the best solution and should be moved to a global
				// (or higher-level) cleanup function.
				ss.CloseFileWriter()
				// Couldn't reconnect to the server, bailout.
				log.Fatalf("error connecting to API stream: %v\n", err)
			}
			log.Println("connected to the API stream.")
		}
		if streamResponse == nil {
			continue
		}
		if ss.streamId != streamResponse.StreamId {
			log.Printf("streamId: %v\n", streamResponse.StreamId)
		}
		ss.streamId = streamResponse.StreamId
		if ss.showStats {
			metrics.setStreamId(ss.streamId)
		}

		switch streamResponse.Response.(type) {
		case *stellarstation.SatelliteStreamResponse_ReceiveTelemetryResponse:
			telemetryResponse := streamResponse.GetReceiveTelemetryResponse()
			if telemetryResponse == nil {
				break
			}
			planId := telemetryResponse.PlanId
			if ss.showStats {
				metrics.setPlanId(planId)
			}
			for _, telemetry := range telemetryResponse.Telemetry {
				if telemetry == nil {
					break
				}
				telemetryData := telemetry.Data
				log.Debug("received data: streamId: %v, planId: %s, groundStationId: %s, framing type: %s, size: %d bytes\n", ss.streamId, planId, ss.groundStationId, telemetry.Framing, len(telemetryData))
				if ss.showStats {
					metrics.collectTelemetry(telemetry)
				}
				if ss.correctOrder {
					go func() {
						ss.mu.Lock()
						defer ss.mu.Unlock()
						pq.Push(telemetry)
					}()
				} else {
					ss.receiveChan <- telemetryData
					if ss.telemetryFileWriter != nil {
						if _, err := ss.telemetryFileWriter.Write(telemetryData); err != nil {
							panic(err)
						}
					}
				}
			}
			// Send ack & update telemetryMessageAckId in case we need to resume from disconnects
			telemetryMessageAckId = telemetryResponse.MessageAckId
			ss.ackReceivedTelemetry(telemetryResponse.MessageAckId)

			// A telemetryResponse containing one telemetry message with a size of zero indicates the stream END message.
			if ss.enableAutoClose && len(telemetryResponse.Telemetry) == 1 && len(telemetryResponse.Telemetry[0].Data) == 0 {
				streamEndDetected = true
			}
		case *stellarstation.SatelliteStreamResponse_StreamEvent:
			if streamResponse.GetStreamEvent() == nil || streamResponse.GetStreamEvent().GetPlanMonitoringEvent() == nil {
				break
			}
			streamEvent := streamResponse.GetStreamEvent()
			monitoringEvent := streamEvent.GetPlanMonitoringEvent()
			planId := monitoringEvent.PlanId

			if ss.isVerbose {
				if gsState := monitoringEvent.GetGroundStationState(); gsState != nil {
					if antennaState := gsState.Antenna; antennaState != nil && antennaState.Azimuth != nil && antennaState.Elevation != nil {
						log.Verbose("planId: %v, azimuth: %v, elevation: %v\n", planId, antennaState.Azimuth.Measured, antennaState.Elevation.Measured)
					}

					if receiverState := gsState.Receiver; receiverState != nil {
						log.Verbose("central frequency (MHz): %.2f\n", float64(receiverState.CenterFrequencyHz)/1e6)
					}
				}
			}
		}
	}
}

func (ss *satelliteStream) openStream(resumeStreamMessageAckId string) error {
	conn, err := apiclient.Dial()
	if err != nil {
		return err
	}

	client := stellarstation.NewStellarStationServiceClient(conn)

	stream, err := client.OpenSatelliteStream(context.Background())
	if err != nil {
		conn.Close()
		return err
	}

	satelliteStreamRequest := stellarstation.SatelliteStreamRequest{
		AcceptedFraming:          ss.acceptedFraming,
		SatelliteId:              ss.satelliteId,
		StreamId:                 ss.streamId,
		ResumeStreamMessageAckId: resumeStreamMessageAckId,
		EnableFlowControl:        true,
	}

	if ss.planId != "" {
		satelliteStreamRequest.PlanId = ss.planId
	}

	if ss.groundStationId != "" {
		satelliteStreamRequest.GroundStationId = ss.groundStationId
	}

	err = stream.Send(&satelliteStreamRequest)
	if err != nil {
		conn.Close()
		return err
	}

	ss.conn = conn
	ss.stream = stream
	if ss.streamId != "" {
		log.Verbose("streamId: %v\n", ss.streamId)
	}

	return nil
}

func (ss *satelliteStream) start() (func(), error) {
	log.SetDebug(ss.isDebug)
	log.SetVerbose(ss.isVerbose)

	// metric collector for data rate, total received size, etc
	if ss.showStats {
		if ss.isVerbose || ss.isDebug {
			metrics = *NewMetricsCollector(log.PrintfRawLn)
			metrics.StartStatsEmitScheduler(2000)
		} else {
			metrics = *NewMetricsCollector(log.LastLine)
			metrics.StartStatsEmitScheduler(500)
		}
	}

	err := ss.openStream("")
	if err != nil {
		return nil, err
	}
	go ss.receiveLoop()

	// return a cleanup function to exec on exit
	cleanup := func() {
		if ss.showStats {
			metrics.logReport()
		}
		ss.CloseFileWriter()
	}
	return cleanup, nil
}
