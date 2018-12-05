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

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/flag"
	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/satellite/tle"
)

var (
	addTLEUse   = util.Normalize("add-tle [Satellite ID] [Line1] [Line2]")
	addTLEShort = util.Normalize("Adds a TLE to a satellite.")
	addTLELong  = util.Normalize("Adds a TLE to a satellite. TLE lines need to be quoted.")
)

// Create add-tle command.
func NewAddTLECommand() *cobra.Command {
	outputFormatFlags := flag.NewOutputFormatFlags()
	flags := flag.NewFlagSet(outputFormatFlags)

	command := &cobra.Command{
		Use:   addTLEUse,
		Short: addTLEShort,
		Long:  addTLELong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return fmt.Errorf("accepts 3 arg(s), received %d", len(args))
			}

			if err := flags.ValidateAll(); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			p := outputFormatFlags.ToPrinter()

			o := &tle.AddTLEOptions{
				Printer:     p,
				SatelliteID: args[0],
				Line1:       args[1],
				Line2:       args[2],
			}

			tle.AddTLE(o)
		},
	}

	// Add flags to the command.
	flags.AddAllFlags(command)

	return command
}
