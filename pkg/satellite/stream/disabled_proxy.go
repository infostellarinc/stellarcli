// Copyright © 2020 Infostellar, Inc.
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
	"net"
	"sync"
)

type noProxy struct {
	recvConn      net.PacketConn
	sendConn      net.Conn
	recvCloseChan chan struct{}
	sendCloseChan chan struct{}

	stream     SatelliteStream
	streamChan chan []byte

	closeWg sync.WaitGroup
}

// Create a connection without using a proxy.
func NewConnectionWithoutProxy() (Proxy, error) {
	streamChan := make(chan []byte)

	p := &noProxy{
		sendCloseChan: make(chan struct{}),
		recvCloseChan: make(chan struct{}),
		streamChan:    streamChan,
		closeWg:       sync.WaitGroup{},
	}

	return p, nil
}

// Start listening for packets to send to the satellite and sending back received packets.
func (p *noProxy) Start(o *SatelliteStreamOptions) (func(), error) {

	var err error
	var cleanup func()
	p.stream, cleanup, err = OpenSatelliteStream(o, p.streamChan)
	if err != nil {
		return cleanup, err
	}

	go p.serve()

	return cleanup, nil
}

func (p *noProxy) serve() {
	for {
		select {
		case <-p.streamChan:
		}
	}
}

// Close the proxy.
func (p *noProxy) Close() error {
	// Close the API stream.
	close(p.streamChan)
	p.stream.Close()

	return nil
}
