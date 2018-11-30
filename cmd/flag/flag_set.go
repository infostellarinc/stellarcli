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

import "github.com/spf13/cobra"

// FlagSet is the collection of flags to be set to the command.
type FlagSet struct {
	flags []Flag
}

// Create a new flag collection.
func NewFlagSet(args ...Flag) *FlagSet {
	s := &FlagSet{}
	for _, f := range args {
		s.flags = append(s.flags, f)
	}
	return s
}

// Add flags to the command.
func (s *FlagSet) AddAllFlags(cmd *cobra.Command) {
	for _, f := range s.flags {
		f.AddFlags(cmd)
	}
}

// Validate all flags.
func (s *FlagSet) ValidateAll() error {
	for _, flag := range s.flags {
		if err := flag.Validate(); err != nil {
			return err
		}
	}

	return nil
}
