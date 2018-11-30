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
	"time"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/util"
)

var (
	// Default time range used when end time is not specified.
	durationInDays uint8 = 31
	// Maximum value of duration in days.
	maxDurationInDays uint8 = 31
	// Time format used to parse "after" and "before" flags.
	timeFormat = "2006-01-02 15:04:05"
)

type PassRangeFlags struct {
	AOSAfter       time.Time
	AOSBefore      time.Time
	DurationInDays uint8

	// Maximum duration in days used in the validation.
	MaxDurationInDays uint8

	flgAOSAfter  string
	flgAOSBefore string
}

// Add flags to the command.
func (f *PassRangeFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.flgAOSAfter, "aos-after", "a", "",
		`The start time (UTC) of the range of plans to list (inclusive). Example: "2006-01-02 15:04:00 (default current time"`)
	cmd.Flags().StringVarP(&f.flgAOSBefore, "aos-before", "b", "",
		`The end time (UTC) of the range of plans to list (exclusive). Example: "2006-01-02 15:14:00" `+
			fmt.Sprintf("(default aos-after + %d days", f.DurationInDays))
	cmd.Flags().Uint8VarP(&f.DurationInDays, "duration", "d", f.DurationInDays,
		fmt.Sprintf("Duration of the range of plans to list (1-%v), in days. Duration will be ignored when aos-before is specified.",
			f.MaxDurationInDays))
}

// Validate flag values.
func (f *PassRangeFlags) Validate() error {
	if f.DurationInDays == 0 || f.DurationInDays > f.MaxDurationInDays {
		return fmt.Errorf("invalid value of duration: %v. Expected value: 1-%v", f.DurationInDays, f.MaxDurationInDays)
	}

	// Validate and set AOSAfter when it is provided.
	f.AOSAfter = time.Now()
	if len(f.flgAOSAfter) > 0 {
		aosAfter, err := util.ParseDateTime(f.flgAOSAfter)
		if err != nil {
			return err
		}
		f.AOSAfter = aosAfter
	}

	// Validate and setAOSBefore when it is provided.
	f.AOSBefore = f.AOSAfter.AddDate(0, 0, int(f.DurationInDays))
	if len(f.flgAOSBefore) > 0 {
		aosBefore, err := util.ParseDateTime(f.flgAOSBefore)
		if err != nil {
			return err
		}
		f.AOSBefore = aosBefore
	}

	if f.AOSAfter.After(f.AOSBefore) {
		return fmt.Errorf("aos-before (%v) must be after aos-after (%v)",
			f.AOSAfter.Format(timeFormat), f.AOSBefore.Format(timeFormat))
	}

	return nil
}

// Create a new PassRangeFlags with default values set.
func NewPassRangeFlags() *PassRangeFlags {
	return &PassRangeFlags{
		DurationInDays:    durationInDays,
		MaxDurationInDays: maxDurationInDays,
	}
}
