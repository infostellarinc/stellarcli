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
	"github.com/infostellarinc/stellarcli/pkg/groundstation/plan"
)

var (
	listPlansUse   = util.Normalize("list-plans [Ground Station ID]")
	listPlansShort = util.Normalize("Lists plans on a ground station.")
	listPlansLong  = util.Normalize(
		`Lists plans on a ground station. Plans having AOS between the given time range are returned.
		When run with default flags, plans in the next 31 days are returned.`)
)

// Create list-plans command.
func NewListPlansCommand() *cobra.Command {
	passRangeFlags := flag.NewPassRangeFlags()
	outputFormatFlags := flag.NewOutputFormatFlags()
	verboseFlag := flag.NewVerboseFlags()
	flags := flag.NewFlagSet(passRangeFlags, outputFormatFlags, verboseFlag)

	command := &cobra.Command{
		Use:   listPlansUse,
		Short: listPlansShort,
		Long:  listPlansLong,
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
			o := &plan.ListOptions{
				Printer:   p,
				ID:        args[0],
				AOSAfter:  &passRangeFlags.AOSAfter,
				AOSBefore: &passRangeFlags.AOSBefore,
				IsVerbose: verboseFlag.IsVerbose,
			}

			plan.ListPlans(o)
		},
	}

	// Add flags to the command.
	flags.AddAllFlags(command)

	return command
}
