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
	"log"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff"
	"google.golang.org/grpc"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
)

const (
	OPEN   uint32 = 0
	CLOSED uint32 = 1
)

const MaxElapsedTime = 60 * time.Second

type SatelliteStreamOptions struct {
	AcceptedFraming []stellarstation.Framing
	AcceptedPlanId  []string
	SatelliteID     string
	IsVerbose       bool
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
	isVerbose      bool
	acceptedPlanId []string
}

// OpenSatelliteStream opens a stream to a satellite over the StellarStation API.
func OpenSatelliteStream(o *SatelliteStreamOptions, recvChan chan<- []byte) (SatelliteStream, error) {
	s := &satelliteStream{
		acceptedFraming:    o.AcceptedFraming,
		satelliteId:        o.SatelliteID,
		streamId:           "",
		recvChan:           recvChan,
		state:              OPEN,
		recvLoopClosedChan: make(chan struct{}),
		isVerbose:          o.IsVerbose,
		acceptedPlanId:     o.AcceptedPlanId,
	}

	err := s.start()

	return s, err
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
		ss.streamId = res.StreamId

		switch res.Response.(type) {
		case *stellarstation.SatelliteStreamResponse_ReceiveTelemetryResponse:
			planId := res.GetReceiveTelemetryResponse().PlanId
			if len(ss.acceptedPlanId) != 0 && !util.Contains(ss.acceptedPlanId, planId) {
				break
			}

			payload := res.GetReceiveTelemetryResponse().Telemetry.Data
			ss.recvChan <- payload
		case *stellarstation.SatelliteStreamResponse_StreamEvent:
			planId := res.GetStreamEvent().GetPlanMonitoringEvent().PlanId
			if len(ss.acceptedPlanId) != 0 && !util.Contains(ss.acceptedPlanId, planId) {
				break
			}

			if ss.isVerbose {
				if gsState := res.GetStreamEvent().GetPlanMonitoringEvent().GetGroundStationState(); gsState != nil {
					if a := gsState.Antenna; a != nil {
						log.Printf("planId: %v, azimuth: %v, elevation: %v\n", planId, a.Azimuth.Measured, a.Elevation.Measured)
					}

					if rcv := gsState.Receiver; rcv != nil {
						log.Printf("central frequency (MHz): %.2f\n", float64(gsState.Receiver.CenterFrequencyHz)/1e6)
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
		AcceptedFraming: ss.acceptedFraming,
		SatelliteId:     ss.satelliteId,
		StreamId:        ss.streamId,
	}

	err = stream.Send(&req)
	if err != nil {
		conn.Close()
		return err
	}

	ss.conn = conn
	ss.stream = stream
	return nil
}

func (ss *satelliteStream) start() error {
	err := ss.openStream()
	if err != nil {
		return err
	}
	go ss.recvLoop()

	return nil
}
