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

package satellite

import (
	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/util"
)

var (
	satelliteUse   = util.Normalize("satellite")
	satelliteShort = util.Normalize("Commands for working with satellites")
)

// Create ground station command.
func NewSatelliteCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     satelliteUse,
		Aliases: []string{"sat"},
		Short:   satelliteShort,
	}

	command.AddCommand(NewCancelPlanCommand())
	command.AddCommand(NewListAvailablePassesCommand())
	command.AddCommand(NewListPlansCommand())
	command.AddCommand(NewOpenStreamCommand())
	command.AddCommand(NewReservePassCommand())

	return command
}
