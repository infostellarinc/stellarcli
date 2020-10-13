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
	"github.com/spf13/cobra"
	"time"
)

var (
	defaultEnableAutoClose = false
	defaultAutoCloseDelay  = 5 * time.Second
	defaultAutoCloseTime   = "2006-01-02'T'15:04:05Z"
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
		"The duration to wait before ending the stream with no more data incoming.")
	cmd.Flags().StringVarP(&f.AutoCloseTime, "auto-close-time", "", defaultAutoCloseTime,
		"The datetime after which auto-closing will be enabled.")

}

// Validate flag values.
func (f *OpenStreamFlag) Validate() error {
	return nil
}

// Create a new OpenStreamFlag with default values set.
func NewOpenStreamFlag() *OpenStreamFlag {
	return &OpenStreamFlag{}
}
