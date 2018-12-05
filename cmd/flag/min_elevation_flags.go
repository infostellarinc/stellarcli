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

	"github.com/spf13/cobra"
)

var (
	// Default min elevation.
	minElevation = 10.0
)

type MinElevationFlags struct {
	MinElevation float64
}

// Add flags to the command.
func (f *MinElevationFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Float64VarP(&f.MinElevation, "min-elevation", "", minElevation,
		"The minimum elevation of passes. Passes are listed having max elevation greater than the minimum elevation.")
}

// Validate flag values.
func (f *MinElevationFlags) Validate() error {
	if f.MinElevation < 0 || f.MinElevation > 90 {
		return fmt.Errorf("invalid value of min elevation: %v. Expected value: 0-90", f.MinElevation)
	}

	return nil
}

// Create a new MinElevationFlags with default values set.
func NewMinElevationFlags() *MinElevationFlags {
	return &MinElevationFlags{
		MinElevation: minElevation,
	}
}
