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

package cmd

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/pkg/stream"
)

var (
	mode       string
	listenHost string
	listenPort uint16
	sendHost   string
	sendPort   uint16
)

// openStreamCmd represents the openStream command
var openStreamCmd = &cobra.Command{
	Use:   "open-stream [satellite-id]",
	Short: "Opens a proxy to stream packets to and from a satellite.",
	Long: `Opens a proxy to stream packets to and from a satellite. Currently only
UDP is supported. Packets received by the proxy will be sent with the specified framing to
the satellite and any incoming packets will be returned as is.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recvAddr := listenHost + ":" + strconv.Itoa(int(listenPort))
		sendAddr := sendHost + ":" + strconv.Itoa(int(sendPort))

		p, err := stream.NewUDPProxy(recvAddr, sendAddr, args[0])
		if err != nil {
			log.Fatalf("Could not open UDP proxy: %v\n", err)
		}

		c := make(chan os.Signal)

		err = p.Start()
		if err != nil {
			log.Fatalf("Could not start UDP proxy: %v\n", err)
		}
		<-c
		p.Close()
	},
}

func init() {
	satelliteCmd.AddCommand(openStreamCmd)

	openStreamCmd.Flags().StringVarP(&mode, "mode", "m", "udp", "The proxy mode to use. One of [udp].")
	openStreamCmd.Flags().StringVar(&listenHost, "listen-host", "127.0.0.1", "The host to listen for packets on.")
	openStreamCmd.Flags().Uint16Var(&listenPort, "listen-port", 6000, "The port to listen for packets on.")
	openStreamCmd.Flags().StringVar(&sendHost, "send-host", "127.0.0.1", "The host to send packets to. Only used by udp.")
	openStreamCmd.Flags().Uint16Var(&sendPort, "send-port", 6001, "The port to send packets on. Only used by udp.")
}
