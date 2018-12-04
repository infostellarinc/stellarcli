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
	uwDurationInDays uint16 = 31
	// Maximum value of duration in days.
	uwMaxDurationInDays uint16 = 365
)

type UWRangeFlags struct {
	StartTime      time.Time
	EndTime        time.Time
	DurationInDays uint16

	// Maximum duration in days used in the validation.
	MaxDurationInDays uint16

	flgStartTime string
	flgEndTime   string
}

// Add flags to the command.
func (f *UWRangeFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.flgStartTime, "start-time", "s", "",
		`The start time (UTC) of the range of unavailability windows to list (inclusive).
			Example: "2006-01-02 15:04:00 (default current time"`)
	cmd.Flags().StringVarP(&f.flgEndTime, "end-time", "e", "",
		`The end time (UTC) of the range of unavailability windows to list (exclusive).
			Example: "2006-01-02 15:14:00" `+fmt.Sprintf("(default start-time + %d days", f.DurationInDays))
	cmd.Flags().Uint16VarP(&f.DurationInDays, "duration", "d", f.DurationInDays,
		fmt.Sprintf("Duration of the range of plans to list (1-%v), in days. Duration will be ignored when end-time is specified.",
			f.MaxDurationInDays))
}

// Validate flag values.
func (f *UWRangeFlags) Validate() error {
	if f.DurationInDays == 0 || f.DurationInDays > f.MaxDurationInDays {
		return fmt.Errorf("invalid value of duration: %v. Expected value: 1-%v", f.DurationInDays, f.MaxDurationInDays)
	}

	// Validate and set StartTime when it is provided.
	f.StartTime = time.Now()
	if f.flgStartTime != "" {
		startTime, err := util.ParseDateTime(f.flgStartTime)
		if err != nil {
			return err
		}
		f.StartTime = startTime
	}

	// Validate and set EndTime when it is provided.
	f.EndTime = f.StartTime.AddDate(0, 0, int(f.DurationInDays))
	if f.flgEndTime != "" {
		endTime, err := util.ParseDateTime(f.flgEndTime)
		if err != nil {
			return err
		}
		f.EndTime = endTime
	}

	if f.StartTime.After(f.EndTime) {
		return fmt.Errorf("aos-before (%v) must be after aos-after (%v)",
			f.StartTime.Format(timeFormat), f.EndTime.Format(timeFormat))
	}

	return nil
}

// Create a new UWRangeFlags with default values set.
func NewUWRangeFlags() *UWRangeFlags {
	return &UWRangeFlags{
		DurationInDays:    uwDurationInDays,
		MaxDurationInDays: uwMaxDurationInDays,
	}
}
