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

type SatelliteStream interface {
	Start()
	Send(payload []byte) error

	io.Closer
}

type satelliteStream struct {
	satelliteId string
	stream      stellarstation.StellarStationService_OpenSatelliteStreamClient
	conn        *grpc.ClientConn

	recvChan           chan []byte
	recvLoopClosedChan chan struct{}

	state uint32
}

// NewSatelliteStream opens a stream to a satellite over the StellarStation API.
func NewSatelliteStream(satelliteId string, recvChan chan []byte) (SatelliteStream, error) {
	conn, err := apiclient.Dial()
	if err != nil {
		return nil, err
	}

	client := stellarstation.NewStellarStationServiceClient(conn)

	stream, err := client.OpenSatelliteStream(context.Background())
	if err != nil {
		conn.Close()
		return nil, err
	}

	req := stellarstation.SatelliteStreamRequest{
		SatelliteId: satelliteId,
	}

	err = stream.Send(&req)
	if err != nil {
		conn.Close()
		return nil, err
	}

	s := &satelliteStream{
		satelliteId:        satelliteId,
		stream:             stream,
		conn:               conn,
		recvChan:           recvChan,
		state:              OPEN,
		recvLoopClosedChan: make(chan struct{}),
	}
	return s, nil
}

// Start starts listening for data from the satellite.
func (ss *satelliteStream) Start() {
	go ss.recvLoop()
}

// Send sends a packet to the satellite.
func (ss *satelliteStream) Send(payload []byte) error {
	req := stellarstation.SatelliteStreamRequest{
		SatelliteId: ss.satelliteId,
		Request: &stellarstation.SatelliteStreamRequest_SendSatelliteCommandsRequest{
			SendSatelliteCommandsRequest: &stellarstation.SendSatelliteCommandsRequest{
				OutputFraming: stellarstation.Framing_AX25,
				Command:       [][]byte{payload},
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
			if err != io.EOF {
				log.Fatalf("Error reading from API stream: %v\n", err)
			}
		} else {
			payload := res.GetReceiveTelemetryResponse().Telemetry.Data
			ss.recvChan <- payload
		}
	}
}
