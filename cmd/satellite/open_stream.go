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

	"github.com/infostellarinc/stellarcli/cmd/flag"
	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/satellite/stream"
	"github.com/spf13/cobra"
)

var (
	openStreamUse   = util.Normalize("open-stream [satellite-id]")
	openStreamShort = util.Normalize("Opens a stream to transfer packets to and from a satellite.")
	openStreamLong  = util.Normalize(
		`Opens a stream to transfer packets to and from a satellite. TCP and UDP proxies are available
		for bidirectional communication in addition to local file writing for reception only. Packets received
		by the proxy will be sent with the specified framing to the satellite and any incoming packets will be
		returned as is.`)
)

// Create open-stream command.
func NewOpenStreamCommand() *cobra.Command {
	debugFlag := flag.NewDebugFlag()
	correctOrderFlags := flag.NewCorrectOrderFlags()
	framingFlags := flag.NewFramingFlags()
	groundStationIdFlag := flag.NewGroundStationIdFlag()
	openStreamFlag := flag.NewOpenStreamFlag()
	planIdFlag := flag.NewPlanIdFlag()
	proxyFlags := flag.NewProxyFlags()
	verboseFlag := flag.NewVerboseFlags()
	statsFlag := flag.NewStatsFlag()
	writeFileFlag := flag.NewWriteFileFlag()
	flags := flag.NewFlagSet(correctOrderFlags, debugFlag, framingFlags, groundStationIdFlag, openStreamFlag, planIdFlag, proxyFlags, verboseFlag, statsFlag, writeFileFlag)

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
				StreamId:        openStreamFlag.StreamId,
				PlanId:          planIdFlag.PlanId,
				GroundStationId: groundStationIdFlag.GroundStationId,
				IsDebug:         debugFlag.IsDebug,
				IsVerbose:       verboseFlag.IsVerbose,
				ShowStats:       statsFlag.ShowStats,
				TelemetryFile:   writeFileFlag.TelemetryFile,

				CorrectOrder:   correctOrderFlags.CorrectOrder,
				DelayThreshold: correctOrderFlags.DelayThreshold,

				EnableAutoClose: openStreamFlag.EnableAutoClose,
			}

			if proxyFlags.ProxyProtocol == "disabled" && writeFileFlag.FileName == "" {
				log.Println("No proxy or output file set. Streamed data will be discarded")
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			defer close(c)

			cleanup, err := proxy.Start(o)
			if err != nil {
				log.Fatalf("could not start proxy: %v\n", err)
			}

			<-c

			if cleanup != nil {
				cleanup()
			}

		},
	}

	// Add flags to the command.
	flags.AddAllFlags(command)

	return command
}
