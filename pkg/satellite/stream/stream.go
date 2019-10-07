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
	"sort"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/api/support/bundler"
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
	IsDebug         bool
	IsVerbose       bool

	CorrectOrder         bool
	BundleCountThreshold int
	BundleByteThreshold  int
	DelayThreshold       time.Duration
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
	acceptedPlanId []string

	correctOrder         bool
	bundleCountThreshold int
	bundleByteThreshold  int
	delayThreshold       time.Duration
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
		isDebug:            o.IsDebug,
		isVerbose:          o.IsVerbose,
		acceptedPlanId:     o.AcceptedPlanId,

		correctOrder:         o.CorrectOrder,
		bundleCountThreshold: o.BundleCountThreshold,
		bundleByteThreshold:  o.BundleByteThreshold,
		delayThreshold:       o.DelayThreshold,
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

	if ss.isVerbose {
		log.Printf("sent data: size: %d bytes\n", len(payload))
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

	payloadBundler := bundler.NewBundler((*stellarstation.Telemetry)(nil), func(bundle interface{}) {
		telemetries := bundle.([]*stellarstation.Telemetry)
		lessFunc := func(i, j int) bool {
			time1, err := ptypes.Timestamp(telemetries[i].TimeFirstByteReceived)
			if err != nil {
				log.Fatal(err)
			}
			time2, err := ptypes.Timestamp(telemetries[j].TimeFirstByteReceived)
			if err != nil {
				log.Fatal(err)
			}

			return time1.Before(time2)
		}
		if !sort.SliceIsSorted(telemetries, lessFunc) {
			sort.SliceStable(telemetries, lessFunc)
		}

		for _, telemetry := range telemetries {
			ss.recvChan <- telemetry.Data
		}
	})
	payloadBundler.DelayThreshold = ss.delayThreshold
	payloadBundler.BundleCountThreshold = ss.bundleCountThreshold
	payloadBundler.BundleByteThreshold = ss.bundleByteThreshold
	payloadBundler.BufferedByteLimit = 1e9 // 1G

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

			telemetry := res.GetReceiveTelemetryResponse().Telemetry
			payload := telemetry.Data
			if ss.isDebug {
				log.Printf("received data: planId: %s, framing type: %s, size: %d bytes\n", planId, telemetry.Framing, len(payload))
			}
			if ss.correctOrder {
				payloadBundler.Add(telemetry, proto.Size(telemetry))
			} else {
				ss.recvChan <- payload
			}
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
