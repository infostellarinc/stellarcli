//
// Copyright © 2020 Infostellar, Inc.
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
	"os"

	"github.com/spf13/cobra"
)

type WriteFileFlag struct {
	FileName      string
	TelemetryFile *os.File
}

// Add a flag to the command.
func (f *WriteFileFlag) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.FileName, "output-file", "o", "", "[Alpha feature] The file to to write packets to. Creates file if it does not exist; appends file if it exists. (default none)")
}

// Validate flag values.
func (f *WriteFileFlag) Validate() error {
	if f.FileName != "" {
		fo, err := os.OpenFile(f.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.TelemetryFile = fo
		return err
	}
	return nil
}

// Create a new WriteFileFlag with default values set.
func NewWriteFileFlag() *WriteFileFlag {
	return &WriteFileFlag{}
}
