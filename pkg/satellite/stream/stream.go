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
	"context"
	"io"
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

	CorrectOrder   bool
	DelayThreshold time.Duration
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

	state          uint32
	isDebug        bool
	isVerbose      bool
	showStats      bool
	acceptedPlanId []string

	correctOrder   bool
	delayThreshold time.Duration
	flushTimer     *time.Timer
	mu             sync.Mutex
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
		acceptedPlanId:     o.AcceptedPlanId,

		correctOrder:   o.CorrectOrder,
		delayThreshold: o.DelayThreshold,
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

	return nil
}

func (ss *satelliteStream) recvLoop() {
	// Initialize exponential back off settings.
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = MaxElapsedTime

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
				err := ss.openStream()
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
				// Couldn't reconnect to the server, bailout.
				log.Fatalf("error connecting to API stream: %v\n", err)
			}
			log.Println("connected to the API stream.")
		}
		if res == nil {
			continue
		}
		if ss.streamId != res.StreamId {
			log.Verbose("streamId: %v\n", res.StreamId)
		}
		ss.streamId = res.StreamId
		if ss.showStats {
			metrics.setStreamId(ss.streamId)
		}

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

			telemetry := telResponse.Telemetry
			if telemetry == nil {
				break
			}
			payload := telemetry.Data
			log.Debug("received data: streamId: %v, planId: %s, index: %d, framing type: %s, size: %d bytes\n", ss.streamId, planId, telemetry.Index, telemetry.Framing, len(payload))
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

				// send ack
				if telResponse.Telemetry.Index > 0 {
					req := stellarstation.SatelliteStreamRequest{
						SatelliteId: ss.satelliteId,
						Request: &stellarstation.SatelliteStreamRequest_TelemetryReceivedAck{
							TelemetryReceivedAck: &stellarstation.ReceiveTelemetryAck{
								PlanId:            planId,
								TelemetryIndex:    telResponse.Telemetry.Index,
								ReceivedTimestamp: timestampNow(),
							},
						},
					}
					log.Info("sending ack index: %v", telResponse.Telemetry.Index)
					ss.stream.Send(&req)
				}
			}
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

func (ss *satelliteStream) openStream() error {
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
		AcceptedFraming:      ss.acceptedFraming,
		SatelliteId:          ss.satelliteId,
		StreamId:             ss.streamId,
		ResumeTelemetryIndex: 1,
		EnableFlowControl:    true,
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

	err := ss.openStream()
	if err != nil {
		return nil, err
	}
	go ss.recvLoop()

	// return a cleanup function to exec on exit
	cleanup := func() {
		if ss.showStats {
			metrics.logReport()
		}
	}
	return cleanup, nil
}
