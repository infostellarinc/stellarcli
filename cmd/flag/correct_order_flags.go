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

	"github.com/infostellarinc/go-stellarstation/api/v1"
)

var (
	// Default correct order.
	defaultCorrectOrder = false

	defaultBundleCountThreshold               = 20
	defaultBundleByteThreshold  int           = 10e6 // 1M
	defaultDelayThreshold       time.Duration = 500 * time.Millisecond
)

type CorrectOrderFlags struct {
	CorrectOrder bool

	BundleCountThreshold int
	BundleByteThreshold  int
	DelayThreshold       time.Duration
}

// Add flags to the command.
func (f *CorrectOrderFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().DurationVarP(&f.DelayThreshold, "delay-threshold", "", defaultDelayThreshold,
		"The maximum amount of time that packets remain in the sorting pool.")
	cmd.Flags().IntVarP(&f.BundleCountThreshold, "count-threshold", "", defaultBundleCountThreshold,
		"The maximum number of packets that will be sorted.")
	cmd.Flags().IntVarP(&f.BundleByteThreshold, "byte-threshold", "", defaultBundleByteThreshold,
		"The maximum number of bytes of packets that will be sorted.")
	cmd.Flags().BoolVarP(&f.CorrectOrder, "correct-order", "", defaultCorrectOrder,
		"Reordering packets by time_first_byte_received when set to true.")
}

// Validate flag values.
func (f *CorrectOrderFlags) Validate() error {
	return nil
}

// Create a new FramingFlags with default values set.
func NewCorrectOrderFlags() *CorrectOrderFlags {
	for value := range v1.Framing_value {
		availableFramings = append(availableFramings, value)
	}

	return &CorrectOrderFlags{
		CorrectOrder: defaultCorrectOrder,
	}
}
