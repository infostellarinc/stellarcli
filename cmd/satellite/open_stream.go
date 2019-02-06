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

package satellite

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/satellite/stream"
)

var (
	openStreamUse   = util.Normalize("open-stream [satellite-id]")
	openStreamShort = util.Normalize("Opens a proxy to stream packets to and from a satellite.")
	openStreamLong  = util.Normalize(
		`Opens a proxy to stream packets to and from a satellite. Currently only
		UDP is supported. Packets received by the proxy will be sent with the specified framing to
		the satellite and any incoming packets will be returned as is.`)
)

var (
	mode       string
	listenHost string
	listenPort uint16
	sendHost   string
	sendPort   uint16
)

// Create open-stream command.
func NewOpenStreamCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   openStreamUse,
		Short: openStreamShort,
		Long:  openStreamLong,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			recvAddr := listenHost + ":" + strconv.Itoa(int(listenPort))
			sendAddr := sendHost + ":" + strconv.Itoa(int(sendPort))

			switch mode {
			case "udp":
				p, err := stream.StartUDPProxy(recvAddr, sendAddr, args[0])
				if err != nil {
					log.Fatalf("Could not open UDP proxy: %v\n", err)
				}
				defer p.Close()
			case "tcp":
				t, err := stream.StartTCPProxy(recvAddr, args[0])
				if err != nil {
					log.Fatalf("Could not open TCP proxy: %v\n", err)
				}
				defer t.Close()
			default:
				log.Fatalf("Unsupported proxy mode: %v\n", mode)
			}

			c := make(chan os.Signal)
			<-c
		},
	}

	command.Flags().StringVarP(&mode, "mode", "m", "udp", "The proxy mode to use. One of [udp,tcp].")
	command.Flags().StringVar(&listenHost, "listen-host", "127.0.0.1", "The host to listen for packets on.")
	command.Flags().Uint16Var(&listenPort, "listen-port", 6000, "The port stellar listens for packets on. Packets on this port will be sent to the satellite.")
	command.Flags().StringVar(&sendHost, "send-host", "127.0.0.1", "The host to send packets to. Only used by udp.")
	command.Flags().Uint16Var(&sendPort, "send-port", 6001, "The port stellar sends packets to. Packets from the satellite will be sent to this port.")

	return command
}
