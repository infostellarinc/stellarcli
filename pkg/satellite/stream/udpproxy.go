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
	"log"
	"net"
	"sync"
	"time"
)

type UDPProxy struct {
	recvConn net.PacketConn
	sendConn net.Conn
	stream   SatelliteStream

	recvBuf       []byte
	recvCloseChan chan struct{}

	sendChan      chan []byte
	sendCloseChan chan struct{}

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
		rc.Close()
		return nil, err
	}

	sc, err := net.Dial("udp", o.SendAddr)
	if err != nil {
		rc.Close()
		sc.Close()
		return nil, err
	}

	sendChan := make(chan []byte)

	p := &UDPProxy{
		recvConn:      rc,
		sendConn:      sc,
		sendChan:      sendChan,
		recvBuf:       make([]byte, 1024*1024),
		sendCloseChan: make(chan struct{}),
		recvCloseChan: make(chan struct{}),
		closeWg:       sync.WaitGroup{},
	}

	return p, nil
}

// Start listening for packets to send to the satellite and sending back received packets.
func (p *UDPProxy) Start(o *SatelliteStreamOptions) error {

	var err error
	p.stream, err = OpenSatelliteStream(o, p.sendChan)
	if err != nil {
		return err
	}

	p.closeWg.Add(2)
	go p.sendLoop()
	go p.recvLoop()

	return nil
}

// Close the proxy.
func (p *UDPProxy) Close() error {

	close(p.sendChan)

	close(p.recvCloseChan)
	close(p.sendCloseChan)

	p.closeWg.Wait()

	p.sendConn.Close()
	p.recvConn.Close()
	p.stream.Close()

	return nil
}

func (p *UDPProxy) recvLoop() {
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

func (p *UDPProxy) sendLoop() {
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
