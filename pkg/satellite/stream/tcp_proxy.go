// Copyright © 2019 Infostellar, Inc.
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
	"net"
	"time"

	log "github.com/infostellarinc/stellarcli/pkg/logger"
)

type tcpProxy struct {
	listener     net.Listener
	connected    chan net.Conn
	disconnected chan net.Conn

	stream      SatelliteStream
	streamChan  chan []byte
	commandChan chan []byte
}

type TCPProxyOptions struct {
	Addr string
}

// Create a UDPProxy.
func NewTCPProxy(o *TCPProxyOptions) (Proxy, error) {
	listener, err := net.Listen("tcp", o.Addr)
	if err != nil {
		log.Fatalf("cannot listen: %v:", err)
	}

	p := &tcpProxy{
		listener:     listener,
		connected:    make(chan net.Conn),
		disconnected: make(chan net.Conn),
		streamChan:   make(chan []byte),
		commandChan:  make(chan []byte),
	}

	return p, nil
}

// Start listening for packets to send to the satellite and sending back received packets.
func (p *tcpProxy) Start(o *SatelliteStreamOptions) error {
	var err error
	p.stream, err = OpenSatelliteStream(o, p.streamChan)
	if err != nil {
		log.Fatalf("failed to connect to StellarStation: %v:", err)
	}

	go p.serve()

	go func() {
		for {
			conn, err := p.listener.Accept()
			if err != nil {
				log.Printf("failed to accept connection: %v", err)
				return
			}
			log.Println("accepted a new connection.")

			go p.handleConn(conn)
		}

	}()

	return nil
}

// Close the proxy.
func (p *tcpProxy) Close() error {
	p.listener.Close()

	return nil
}

// Sends packets received from Satellite to all clients.
func (p *tcpProxy) serve() {
	conns := make(map[net.Conn]bool)

	for {
		select {
		case conn := <-p.connected:
			conns[conn] = true
			log.Println("connected to a new client:", conn.RemoteAddr().String())
			log.Println("connected clients:", len(conns))
		case conn := <-p.disconnected:
			delete(conns, conn)
			conn.Close()
			log.Println("disconnected the client:", conn.RemoteAddr().String())
			log.Println("connected clients:", len(conns))
		case payload := <-p.streamChan:
			for conn := range conns {
				conn.Write(payload)
			}
		case command := <-p.commandChan:
			p.stream.Send(command)
		}
	}

}
func (p *tcpProxy) handleConn(conn net.Conn) {
	p.connected <- conn

	defer func() {
		p.disconnected <- conn
	}()

	buf := make([]byte, 1024*1024)
	for {
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			} else if !err.(net.Error).Timeout() {
				log.Fatal(err)
				return
			}
			// Pass through timeout error.
		} else {
			p.commandChan <- buf[:n]
		}
	}
}
