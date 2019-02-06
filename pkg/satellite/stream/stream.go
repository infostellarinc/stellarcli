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

	"google.golang.org/grpc"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
)

const (
	OPEN   uint32 = 0
	CLOSED uint32 = 1
)

type SatelliteStreamOptions struct {
	SatelliteID string
}

type SatelliteStream interface {
	Send(payload []byte) error

	io.Closer
}

type satelliteStream struct {
	satelliteId string
	stream      stellarstation.StellarStationService_OpenSatelliteStreamClient
	conn        *grpc.ClientConn
	streamId    string

	recvChan           chan []byte
	recvLoopClosedChan chan struct{}

	state uint32
}

// OpenSatelliteStream opens a stream to a satellite over the StellarStation API.
func OpenSatelliteStream(o *SatelliteStreamOptions, recvChan chan []byte) (SatelliteStream, error) {
	s := &satelliteStream{
		satelliteId:        o.SatelliteID,
		streamId:           "",
		recvChan:           recvChan,
		state:              OPEN,
		recvLoopClosedChan: make(chan struct{}),
	}

	s.start()

	return s, nil
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
	for {
		res, err := ss.stream.Recv()
		if atomic.LoadUint32(&ss.state) == CLOSED {
			// Closed, so just shutdown the loop.
			close(ss.recvLoopClosedChan)
			return
		}
		if err != nil {
			if err == io.EOF {
				// Server closed the stream, try to reconnect.
				err = ss.openStream()
				if err != nil {
					// Couldn't reconnect to the server, bailout.
					log.Fatalf("Error opening API stream: %v\n", err)
				}
			} else {
				log.Fatalf("Error reading from API stream: %v\n", err)
			}
		}

		ss.streamId = res.StreamId

		switch res.Response.(type) {
		case *stellarstation.SatelliteStreamResponse_ReceiveTelemetryResponse:
			payload := res.GetReceiveTelemetryResponse().Telemetry.Data
			ss.recvChan <- payload
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
		SatelliteId: ss.satelliteId,
		StreamId:    ss.streamId,
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
