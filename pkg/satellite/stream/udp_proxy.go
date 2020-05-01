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
	"net"
	"sync"
	"time"

	log "github.com/infostellarinc/stellarcli/pkg/logger"
)

type udpProxy struct {
	recvConn      net.PacketConn
	sendConn      net.Conn
	recvCloseChan chan struct{}
	sendCloseChan chan struct{}

	stream     SatelliteStream
	streamChan chan []byte

	closeWg sync.WaitGroup
}

type UDPProxyOptions struct {
	RecvAddr string
	SendAddr string
}

// Create a UDPProxy.
func NewUDPProxy(o *UDPProxyOptions) (Proxy, error) {
	rc, err := net.ListenPacket("udp", o.RecvAddr)
	if err != nil {
		if rc != nil {
			rc.Close()
		}
		return nil, err
	}

	sc, err := net.Dial("udp", o.SendAddr)
	if err != nil {
		if rc != nil {
			rc.Close()
		}
		if sc != nil {
			sc.Close()
		}
		return nil, err
	}

	streamChan := make(chan []byte)

	p := &udpProxy{
		recvConn:      rc,
		sendConn:      sc,
		sendCloseChan: make(chan struct{}),
		recvCloseChan: make(chan struct{}),
		streamChan:    streamChan,
		closeWg:       sync.WaitGroup{},
	}

	return p, nil
}

// Start listening for packets to send to the satellite and sending back received packets.
func (p *udpProxy) Start(o *SatelliteStreamOptions) error {

	var err error
	p.stream, err = OpenSatelliteStream(o, p.streamChan)
	if err != nil {
		return err
	}

	p.closeWg.Add(2)
	go p.sendLoop()
	go p.recvLoop()

	return nil
}

// Close the proxy.
func (p *udpProxy) Close() error {

	// Close connections used in send/receive loop.
	close(p.recvCloseChan)
	close(p.sendCloseChan)
	p.closeWg.Wait()

	// Close the API stream.
	close(p.streamChan)
	p.stream.Close()

	return nil
}

func (p *udpProxy) recvLoop() {
	defer p.closeWg.Done()

	recvBuf := make([]byte, 1024*1024)

	for {
		select {
		case <-p.recvCloseChan:
			p.recvConn.Close()
			return
		default:
			p.recvConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, _, err := p.recvConn.ReadFrom(recvBuf)
			if err != nil {
				if !err.(net.Error).Timeout() {
					log.Fatalf("error receiving on UDP port: %v\n", err)
				}
			} else {
				p.stream.Send(recvBuf[:n])
			}
		}
	}
}

func (p *udpProxy) sendLoop() {
	defer p.closeWg.Done()
	for {
		select {
		case payload := <-p.streamChan:
			p.sendConn.Write(payload)
		case <-p.sendCloseChan:
			p.sendConn.Close()
			return
		}
	}
}
