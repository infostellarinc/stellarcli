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
	"github.com/spf13/cobra"
	"time"
	"github.com/infostellarinc/stellarcli/pkg/plan"
	"fmt"
)

const (
	// Time format used to parse "after" and "before" flags.
	timeFormat = "2006-01-02 15:04:05"
	// Default time range used when end time is not specified.
	defaultDurationInDays = 31
	// Maximum value of duration in days.
	maxDurationInDays = 31
)

var (
	flgAOSAfter  string
	flgAOSBefore string
	flgDuration  uint8
)

// listGSPlansCmd represents the ground station command
var listGSPlansCmd = &cobra.Command{
	Use:   "list-plans [Ground Station ID]",
	Short: "Lists plans on a ground station.",
	Long:  `Lists plans on a ground station. Plans having AOS between the given time range will be returned.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}

		if flgDuration == 0 || flgDuration > maxDurationInDays {
			return fmt.Errorf("Invalid value of duration: %v. Expected value: 1-%v", flgDuration, maxDurationInDays)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		aosAfter, err := time.Parse(timeFormat, flgAOSAfter)
		if err != nil {
			aosAfter = time.Now()
		}

		aosBefore, err := time.Parse(timeFormat, flgAOSBefore)
		if err != nil {
			aosBefore = aosAfter.AddDate(0, 0, int(flgDuration))
		}

		plan.ListPlans(args[0], aosAfter, aosBefore)
	},
}

func init() {
	gsCmd.AddCommand(listGSPlansCmd)

	listGSPlansCmd.Flags().StringVarP(&flgAOSAfter, "aos-after", "a", "",
		`The start time (UTC) of the range of plans to list (inclusive). Example: "2006-01-02 15:04:00 (default current time"`)
	listGSPlansCmd.Flags().StringVarP(&flgAOSBefore, "aos-before", "b", "",
		`The end time (UTC) of the range of plans to list (exclusive). Example: "2006-01-02 15:14:00" `+
			fmt.Sprintf("(default current time + %d days", defaultDurationInDays))
	listGSPlansCmd.Flags().Uint8VarP(&flgDuration, "duration", "d", defaultDurationInDays,
		fmt.Sprintf("Duration of the range of plans to list (1-%v). Duration will be ignored when aos-before is specified.", maxDurationInDays))
}
