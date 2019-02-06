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
	availableProxy = []string{"udp"}
	// Default listen host
	defaultListenHost = "127.0.0.1"
	// Default listen port
	defaultListenPort uint16 = 6000
	// Default send host
	defaultSendHost = "127.0.0.1"
	// Default send port
	defaultSendPort uint16 = 6001
)

type OpenStreamFlags struct {
	ProxyProtocol string

	ListenHost string
	ListenPort uint16
	SendHost   string
	SendPort   uint16
}

// Add flags to the command.
func (f *OpenStreamFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.ProxyProtocol, "proxy", "", defaultProxyProtocol,
		"Proxy protocol. One of: "+strings.Join(availableFormats, "|"))
	cmd.Flags().StringVar(&f.ListenHost, "listen-host", defaultListenHost,
		"The host to listen for packets on.")
	cmd.Flags().Uint16Var(&f.ListenPort, "listen-port", defaultListenPort,
		"The port stellar listens for packets on. Packets on this port will be sent to the satellite.")
	cmd.Flags().StringVar(&f.SendHost, "send-host", defaultSendHost,
		"The host to send packets to. Only used by udp.")
	cmd.Flags().Uint16Var(&f.SendPort, "send-port", defaultSendPort,
		"The port stellar sends packets to. Packets from the satellite will be sent to this port.")
}

// Validate flag values.
func (f *OpenStreamFlags) Validate() error {
	if !util.Contains(availableProxy, f.ProxyProtocol) {
		return fmt.Errorf("invalid proxy protocol: %v. Expected one of: %v", f.ProxyProtocol,
			strings.Join(availableProxy, "|"))
	}

	return nil
}

// Return a Proxy corresponding to the protocol.
func (f *OpenStreamFlags) ToProxy() stream.Proxy {
	protocol := util.ToLower(f.ProxyProtocol)

	recvAddr := fmt.Sprintf("%s:%d", f.ListenHost, f.ListenPort)
	sendAddr := fmt.Sprintf("%s:%d", f.SendHost, f.SendPort)

	switch protocol {
	case "udp":
		o := &stream.UDPProxyOptions{
			RecvAddr: recvAddr,
			SendAddr: sendAddr,
		}
		p, err := stream.NewUDPProxy(o)
		if err != nil {
			log.Fatalf("Could not open UDP proxy: %v\n", err)
		}
		return p
	}

	log.Fatalf("Unsupported proxy protocol: %v", protocol)
	return nil
}

// Create a new OpenStreamFlags with default values set.
func NewOpenStreamFlags() *OpenStreamFlags {
	return &OpenStreamFlags{
		ProxyProtocol: defaultProxyProtocol,

		ListenHost: defaultListenHost,
		ListenPort: defaultListenPort,
		SendHost:   defaultSendHost,
		SendPort:   defaultSendPort,
	}
}
