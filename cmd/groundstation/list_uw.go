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

package groundstation

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/flag"
	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/groundstation/uw"
)

var (
	listUWUse   = util.Normalize("list-uw [Ground Station ID]")
	listUWShort = util.Normalize("Lists unavailability windows on a ground station.")
	listUWLong  = util.Normalize(
		`Lists unavailability windows on a ground station. Unavailability windows between the given time range
		are returned.`)
)

// Create list-plans command.
func NewListUWCommand() *cobra.Command {
	uwRangeFlags := flag.NewUWRangeFlags()
	outputFormatFlags := flag.NewOutputFormatFlags()
	flags := flag.NewFlagSet(uwRangeFlags, outputFormatFlags)

	command := &cobra.Command{
		Use:   listUWUse,
		Short: listUWShort,
		Long:  listUWLong,
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
			p := outputFormatFlags.ToPrinter()
			o := &uw.ListUWOptions{
				Printer:   p,
				ID:        args[0],
				StartTime: uwRangeFlags.StartTime,
				EndTime:   uwRangeFlags.EndTime,
			}

			uw.ListUW(o)
		},
	}

	// Add flags to the command.
	flags.AddAllFlags(command)

	return command

}
