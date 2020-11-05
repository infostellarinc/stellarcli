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
	defaultEnableAutoClose = false
	defaultAutoCloseDelay  = 5 * time.Second
)

type OpenStreamFlag struct {
	EnableAutoClose bool
	AutoCloseDelay  time.Duration
	AutoCloseTime   string
	StreamId        string
}

// Add a flag to the command.
func (f *OpenStreamFlag) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.StreamId, "stream-id", "r", "", "The StreamId to resume.")
	cmd.Flags().BoolVarP(&f.EnableAutoClose, "enable-auto-close", "", defaultEnableAutoClose,
		"When set to true, the stream will close after a specified auto close time.")
	cmd.Flags().DurationVarP(&f.AutoCloseDelay, "auto-close-delay", "", defaultAutoCloseDelay,
		"The duration to wait before ending the stream with no more data incoming. Valid time units are \"s\", \"m\". Ex 1m30s. Range 1s to 10m")
	cmd.Flags().StringVarP(&f.AutoCloseTime, "auto-close-time", "", "",
		"The datetime (UTC) after which auto-closing will be enabled. Format 2006-01-02 15:04:05")
}

// Validate flag values.
func (f *OpenStreamFlag) Validate() error {
	return nil
}

// Create a new OpenStreamFlag with default values set.
func NewOpenStreamFlag() *OpenStreamFlag {
	return &OpenStreamFlag{}
}
