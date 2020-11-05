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

package flag

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/satellite/stream"
)

var (
	// Default proxy protocol.
	defaultProxyProtocol = "udp"

	// Supported proxy.
	availableProxy = []string{"udp", "tcp", "disabled"}
	// Default listen host for UDP.
	defaultUDPListenHost = "127.0.0.1"
	// Default listen port for UDP.
	defaultUDPListenPort uint16 = 6000
	// Default send host for UDP.
	defaultUDPSendHost = "127.0.0.1"
	// Default send port for UDP.
	defaultUDPSendPort uint16 = 6001

	// Default listen host for TCP.
	defaultTCPListenHost = "127.0.0.1"
	// Default listen port for TCP.
	defaultTCPListenPort uint16 = 6001
)

type ProxyFlags struct {
	ProxyProtocol string

	UDPListenHost string
	UDPListenPort uint16
	UDPSendHost   string
	UDPSendPort   uint16

	TCPListenHost string
	TCPListenPort uint16
}

// Add flags to the command.
func (f *ProxyFlags) AddFlags(cmd *cobra.Command) {
	// Currently defaults to UDP.
	cmd.Flags().StringVarP(&f.ProxyProtocol, "proxy", "", defaultProxyProtocol,
		"Proxy protocol. One of: "+strings.Join(availableProxy, "|"))

	cmd.Flags().StringVar(&f.UDPListenHost, "listen-host", "", "Deprecated: use udp-listen-host instead.")
	cmd.Flags().Uint16Var(&f.UDPListenPort, "listen-port", 0, "Deprecated: use udp-listen-port instead.")
	cmd.Flags().StringVar(&f.UDPSendHost, "send-host", "", "Deprecated: use udp-send-host instead.")
	cmd.Flags().Uint16Var(&f.UDPSendPort, "send-port", 0, "Deprecated: use udp-send-port instead.")

	cmd.Flags().StringVar(&f.UDPListenHost, "udp-listen-host", defaultUDPListenHost,
		"The host to listen for packets on.")
	cmd.Flags().Uint16Var(&f.UDPListenPort, "udp-listen-port", defaultUDPListenPort,
		"The port stellar listens for packets on. Packets on this port will be sent to the satellite.")
	cmd.Flags().StringVar(&f.UDPSendHost, "udp-send-host", defaultUDPSendHost,
		"The host to send UDP packets to.")
	cmd.Flags().Uint16Var(&f.UDPSendPort, "udp-send-port", defaultUDPSendPort,
		"The port stellar sends UDP packets to. Packets from the satellite will be sent to this port.")

	cmd.Flags().StringVar(&f.TCPListenHost, "tcp-listen-host", defaultTCPListenHost,
		"The host to listen for TCP connection on.")
	cmd.Flags().Uint16Var(&f.TCPListenPort, "tcp-listen-port", defaultTCPListenPort,
		"The port used to communicate with satellite. Clients can receive and send data through the port.")
}

// Validate flag values.
func (f *ProxyFlags) Validate() error {
	if !util.Contains(availableProxy, f.ProxyProtocol) {
		return fmt.Errorf("invalid proxy protocol: %v. Expected one of: %v", f.ProxyProtocol,
			strings.Join(availableProxy, "|"))
	}

	return nil
}

// Return a Proxy corresponding to the protocol.
func (f *ProxyFlags) ToProxy() stream.Proxy {
	protocol := util.ToLower(f.ProxyProtocol)

	switch protocol {
	case "udp":
		recvAddr := fmt.Sprintf("%s:%d", f.UDPListenHost, f.UDPListenPort)
		sendAddr := fmt.Sprintf("%s:%d", f.UDPSendHost, f.UDPSendPort)

		o := &stream.UDPProxyOptions{
			RecvAddr: recvAddr,
			SendAddr: sendAddr,
		}
		p, err := stream.NewUDPProxy(o)
		if err != nil {
			log.Fatalf("could not open UDP proxy: %v\n", err)
		}
		return p
	case "tcp":
		addr := fmt.Sprintf("%s:%d", f.TCPListenHost, f.TCPListenPort)
		o := &stream.TCPProxyOptions{
			Addr: addr,
		}
		p, err := stream.NewTCPProxy(o)
		if err != nil {
			log.Fatalf("could not open TCP proxy: %v\n", err)
		}
		return p
	case "disabled":
		p, err := stream.NewConnectionWithoutProxy()
		if err != nil {
			log.Fatalf("could not open connection: %v\n", err)
		}
		return p
	}

	log.Fatalf("unsupported proxy protocol: %v", protocol)
	return nil
}

// Create a new ProxyFlags with default values set.
func NewProxyFlags() *ProxyFlags {
	return &ProxyFlags{
		ProxyProtocol: defaultProxyProtocol,

		UDPListenHost: defaultUDPListenHost,
		UDPListenPort: defaultUDPListenPort,
		UDPSendHost:   defaultUDPSendHost,
		UDPSendPort:   defaultUDPSendPort,
		TCPListenHost: defaultTCPListenHost,
		TCPListenPort: defaultTCPListenPort,
	}
}
