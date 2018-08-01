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
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type UDPProxy interface {
	Start() error
	io.Closer
}

type udpProxy struct {
	recvConn net.PacketConn
	sendConn net.Conn
	stream   SatelliteStream

	recvBuf       []byte
	recvCloseChan chan struct{}

	sendChan      chan []byte
	sendCloseChan chan struct{}

	closeWg sync.WaitGroup
}

// NewUDPProxy creates a UDPProxy that will listen for packets to send to the satellite and send back
// received packets.
func NewUDPProxy(recvAddr string, sendAddr string, satelliteId string) (UDPProxy, error) {
	rc, err := net.ListenPacket("udp", recvAddr)
	if err != nil {
		return nil, err
	}

	sc, err := net.Dial("udp", sendAddr)
	if err != nil {
		rc.Close()
		return nil, err
	}

	sendChan := make(chan []byte)

	stream, err := NewSatelliteStream(satelliteId, sendChan)
	if err != nil {
		rc.Close()
		sc.Close()
		return nil, err
	}

	p := &udpProxy{
		recvConn:      rc,
		sendConn:      sc,
		stream:        stream,
		sendChan:      sendChan,
		recvBuf:       make([]byte, 1024*1024),
		sendCloseChan: make(chan struct{}),
		recvCloseChan: make(chan struct{}),
		closeWg:       sync.WaitGroup{},
	}
	return p, nil
}

// Start starts the proxy, listening for packets to send to/from the satellite.
func (p *udpProxy) Start() error {
	err := p.stream.Start()
	if err != nil {
		return err
	}

	go p.sendLoop()
	go p.recvLoop()

	return nil
}

// Close closes the proxy.
func (p *udpProxy) Close() error {
	p.stream.Close()

	close(p.recvCloseChan)
	close(p.sendCloseChan)

	p.closeWg.Wait()

	p.sendConn.Close()
	p.recvConn.Close()

	return nil
}

func (p *udpProxy) recvLoop() {
	p.closeWg.Add(1)
	defer p.closeWg.Done()
	for {
		select {
		case <-p.recvCloseChan:
			return
		default:
			p.recvConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, _, err := p.recvConn.ReadFrom(p.recvBuf)
			if err != nil {
				if !err.(net.Error).Timeout() {
					log.Fatalf("Error receiving on UDP port: %v\n", err)
				}
			} else {
				p.stream.Send(p.recvBuf[:n])
			}
		}
	}
}

func (p *udpProxy) sendLoop() {
	p.closeWg.Add(1)
	defer p.closeWg.Done()
	for {
		select {
		case payload := <-p.sendChan:
			p.sendConn.Write(payload)
		case <-p.sendCloseChan:
			return
		}
	}
}
