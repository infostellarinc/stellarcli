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
	"github.com/infostellarinc/stellarcli/cmd/util"
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
	AcceptedPlanId  []string
	SatelliteID     string
	StreamId        string
	IsDebug         bool
	IsVerbose       bool
	ShowStats       bool
	TelemetryFile   *os.File

	CorrectOrder   bool
	DelayThreshold time.Duration

	EnableAutoClose bool
	AutoCloseDelay  time.Duration
	AutoCloseTime   time.Time
}

type SatelliteStream interface {
	Send(payload []byte) error

	io.Closer
}

type satelliteStream struct {
	acceptedFraming []stellarstation.Framing

	satelliteId string
	stream      stellarstation.StellarStationService_OpenSatelliteStreamClient
	conn        *grpc.ClientConn
	streamId    string

	recvChan           chan<- []byte
	recvLoopClosedChan chan struct{}

	telemetryFileWriter *bufio.Writer

	state          uint32
	isDebug        bool
	isVerbose      bool
	showStats      bool
	telemetryFile  *os.File
	acceptedPlanId []string

	correctOrder   bool
	delayThreshold time.Duration
	flushTimer     *time.Timer
	mu             sync.Mutex

	enableAutoClose bool
	autoCloseDelay  time.Duration
	autoCloseTime   time.Time
}

// OpenSatelliteStream opens a stream to a satellite over the StellarStation API.
func OpenSatelliteStream(o *SatelliteStreamOptions, recvChan chan<- []byte) (SatelliteStream, func(), error) {
	s := &satelliteStream{
		acceptedFraming:    o.AcceptedFraming,
		satelliteId:        o.SatelliteID,
		streamId:           o.StreamId,
		recvChan:           recvChan,
		state:              OPEN,
		recvLoopClosedChan: make(chan struct{}),
		isDebug:            o.IsDebug,
		isVerbose:          o.IsVerbose,
		showStats:          o.ShowStats,
		telemetryFile:      o.TelemetryFile,
		acceptedPlanId:     o.AcceptedPlanId,

		correctOrder:   o.CorrectOrder,
		delayThreshold: o.DelayThreshold,

		enableAutoClose: o.EnableAutoClose,
		autoCloseDelay:  o.AutoCloseDelay,
		autoCloseTime:   o.AutoCloseTime,
	}

	cleanup, err := s.start()

	return s, cleanup, err
}

// Send sends a packet to the satellite.
func (ss *satelliteStream) Send(payload []byte) error {
	req := stellarstation.SatelliteStreamRequest{
		SatelliteId: ss.satelliteId,
		Request: &stellarstation.SatelliteStreamRequest_SendSatelliteCommandsRequest{
			SendSatelliteCommandsRequest: &stellarstation.SendSatelliteCommandsRequest{
				Command: [][]byte{payload},
			},
		},
	}

	log.Verbose("sent data: size: %d bytes\n", len(payload))

	return ss.stream.Send(&req)
}

// Close closes the stream.
func (ss *satelliteStream) Close() error {
	atomic.StoreUint32(&ss.state, CLOSED)

	ss.stream.CloseSend()
	ss.conn.Close()

	<-ss.recvLoopClosedChan

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
		req := stellarstation.SatelliteStreamRequest{
			SatelliteId: ss.satelliteId,
			Request: &stellarstation.SatelliteStreamRequest_TelemetryReceivedAck{
				TelemetryReceivedAck: &stellarstation.ReceiveTelemetryAck{
					MessageAckId:      telemetryMessageAckId,
					ReceivedTimestamp: timestampNow(),
				},
			},
		}
		log.Debug("sending ack index: %v", telemetryMessageAckId)
		ss.stream.Send(&req)
	}
}

func (ss *satelliteStream) recvLoop() {
	// Initialize exponential back off settings.
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = MaxElapsedTime

	// Initialize auto close
	receivingBytes := false
	d := time.Now().Add(time.Second * ss.autoCloseDelay)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

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
			ss.recvChan <- telemetry.Data
			if ss.telemetryFileWriter != nil {
				if _, err := ss.telemetryFileWriter.Write(telemetry.Data); err != nil {
					panic(err)
				}
			}
		}
		ss.flushTimer.Reset(ss.delayThreshold)
	})

	for {
		res, err := ss.stream.Recv()
		if atomic.LoadUint32(&ss.state) == CLOSED {
			// Closed, so just shutdown the loop.
			close(ss.recvLoopClosedChan)
			return
		}
		if err != nil {
			log.Println(err)
			log.Println("reconnecting to the API stream.")

			rcErr := backoff.RetryNotify(func() error {
				err := ss.openStream(telemetryMessageAckId)
				if err != nil {
					return err
				}

				response, err := ss.stream.Recv()
				if err != nil {
					return err
				}
				res = response

				return nil
			}, b,
				func(e error, duration time.Duration) {
					log.Printf("%s. Automatically retrying in %v", e, duration)
				})
			if rcErr != nil {
				// This explicit cleanup is not the best solution and should be moved to a global
				// (or higher-level) cleanup function.
				ss.CloseFileWriter()
				// Couldn't reconnect to the server, bailout.
				log.Fatalf("error connecting to API stream: %v\n", err)
			}
			log.Println("connected to the API stream.")
		}
		if res == nil {
			continue
		}
		if ss.streamId != res.StreamId {
			log.Printf("streamId: %v\n", res.StreamId)
		}
		ss.streamId = res.StreamId
		if ss.showStats {
			metrics.setStreamId(ss.streamId)
		}

		select {
		case <-ctx.Done():
			// Close stream if end of data and the current time is after auto close time
			if ss.enableAutoClose && time.Now().UTC().After(ss.autoCloseTime) && !receivingBytes {
				close(ss.recvLoopClosedChan)
				return
			} else {
				// Renew the deadline and check again after it expires
				d := time.Now().Add(time.Second * ss.autoCloseDelay)
				ctx, cancel = context.WithDeadline(context.Background(), d)
			}
		}
		receivingBytes = false

		switch res.Response.(type) {
		case *stellarstation.SatelliteStreamResponse_ReceiveTelemetryResponse:
			telResponse := res.GetReceiveTelemetryResponse()
			if telResponse == nil {
				break
			}
			planId := telResponse.PlanId
			if ss.showStats {
				metrics.setPlanId(planId)
			}
			if len(ss.acceptedPlanId) != 0 && !util.Contains(ss.acceptedPlanId, planId) {
				break
			}

			for _, telemetry := range telResponse.Telemetry {
				if telemetry == nil {
					break
				}
				payload := telemetry.Data
				log.Debug("received data: streamId: %v, planId: %s, framing type: %s, size: %d bytes\n", ss.streamId, planId, telemetry.Framing, len(payload))
				if ss.enableAutoClose {
					receivingBytes = true
				}
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
					ss.recvChan <- payload
					if ss.telemetryFileWriter != nil {
						if _, err := ss.telemetryFileWriter.Write(payload); err != nil {
							panic(err)
						}
					}
				}
			}
			// send ack & update telemetryMessageAckId in case we need to resume from disconnects
			telemetryMessageAckId = telResponse.MessageAckId
			ss.ackReceivedTelemetry(telResponse.MessageAckId)
		case *stellarstation.SatelliteStreamResponse_StreamEvent:
			if res.GetStreamEvent() == nil || res.GetStreamEvent().GetPlanMonitoringEvent() == nil {
				break
			}
			streamEvent := res.GetStreamEvent()
			monitoringEvent := streamEvent.GetPlanMonitoringEvent()
			planId := monitoringEvent.PlanId
			if len(ss.acceptedPlanId) != 0 && !util.Contains(ss.acceptedPlanId, planId) {
				break
			}

			if ss.isVerbose {
				if gsState := monitoringEvent.GetGroundStationState(); gsState != nil {
					if a := gsState.Antenna; a != nil && a.Azimuth != nil && a.Elevation != nil {
						log.Verbose("planId: %v, azimuth: %v, elevation: %v\n", planId, a.Azimuth.Measured, a.Elevation.Measured)
					}

					if rcv := gsState.Receiver; rcv != nil {
						log.Verbose("central frequency (MHz): %.2f\n", float64(rcv.CenterFrequencyHz)/1e6)
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

	req := stellarstation.SatelliteStreamRequest{
		AcceptedFraming:          ss.acceptedFraming,
		SatelliteId:              ss.satelliteId,
		StreamId:                 ss.streamId,
		ResumeStreamMessageAckId: resumeStreamMessageAckId,
		EnableFlowControl:        true,
	}

	err = stream.Send(&req)
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
	go ss.recvLoop()

	// return a cleanup function to exec on exit
	cleanup := func() {
		if ss.showStats {
			metrics.logReport()
		}
		ss.CloseFileWriter()
	}
	return cleanup, nil
}
