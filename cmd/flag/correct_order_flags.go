//
// Copyright © 2019 Infostellar, Inc.
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
	"time"

	"github.com/spf13/cobra"
)

var (
	// Default correct order.
	defaultCorrectOrder                 = false
	defaultDelayThreshold time.Duration = 500 * time.Millisecond
)

type CorrectOrderFlags struct {
	CorrectOrder   bool
	DelayThreshold time.Duration
}

// Add flags to the command.
func (f *CorrectOrderFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().DurationVarP(&f.DelayThreshold, "delay-threshold", "", defaultDelayThreshold,
		"The maximum amount of time that packets remain in the sorting pool.")
	cmd.Flags().BoolVarP(&f.CorrectOrder, "correct-order", "", defaultCorrectOrder,
		"When set to true, packets will be sorted by time_first_byte_received. This feature is alpha quality.")
}

// Validate flag values.
func (f *CorrectOrderFlags) Validate() error {
	return nil
}

// Create a new CorrectOrderFlags with default values set.
func NewCorrectOrderFlags() *CorrectOrderFlags {
	return &CorrectOrderFlags{
		CorrectOrder: defaultCorrectOrder,
	}
}
