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
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/cmd/util"
)

var (
	// Supported framings.
	availableFramings []string
	// Default accepted framing.
	defaultAcceptedFraming []string
)

type FramingFlags struct {
	AcceptedFraming []string
}

// Add flags to the command.
func (f *FramingFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&f.AcceptedFraming, "accepted-framing", defaultAcceptedFraming,
		"Framing type to receive. One of: "+strings.Join(availableFramings, "|"))
}

// Validate flag values.
func (f *FramingFlags) Validate() error {
	for _, framing := range f.AcceptedFraming {
		if !util.Contains(availableFramings, framing) {
			return fmt.Errorf("invalid framing type: %v. Expected one of : %v", framing,
				strings.Join(availableFramings, "|"))
		}
	}

	return nil
}

// Return accepted framing
func (f *FramingFlags) ToProtoAcceptedFraming() []v1.Framing {
	var acceptedFrame []v1.Framing
	for _, framing := range f.AcceptedFraming {
		acceptedFrame = append(acceptedFrame, v1.Framing(v1.Framing_value[framing]))
	}

	return acceptedFrame
}

// Create a new FramingFlags with default values set.
func NewFramingFlags() *FramingFlags {
	for value := range v1.Framing_value {
		availableFramings = append(availableFramings, value)
	}

	return &FramingFlags{
		AcceptedFraming: defaultAcceptedFraming,
	}
}
