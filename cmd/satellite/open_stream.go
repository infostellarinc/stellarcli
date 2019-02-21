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
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/flag"
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

// Create open-stream command.
func NewOpenStreamCommand() *cobra.Command {
	proxyFlags := flag.NewProxyFlags()
	framingFlags := flag.NewFramingFlags()
	flags := flag.NewFlagSet(proxyFlags, framingFlags)

	command := &cobra.Command{
		Use:   openStreamUse,
		Short: openStreamShort,
		Long:  openStreamLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
			}

			if err := flags.ValidateAll(); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			proxy := proxyFlags.ToProxy()
			defer proxy.Close()

			o := &stream.SatelliteStreamOptions{
				SatelliteID:     args[0],
				AcceptedFraming: framingFlags.ToProtoAcceptedFraming(),
			}

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt)
			defer close(c)

			err := proxy.Start(o)
			if err != nil {
				log.Fatalf("Could not start proxy: %v\n", err)
			}

			<-c

		},
	}

	// Add flags to the command.
	flags.AddAllFlags(command)

	return command
}
