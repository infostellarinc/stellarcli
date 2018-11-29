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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/auth"
	"github.com/infostellarinc/stellarcli/cmd/groundstation"
	"github.com/infostellarinc/stellarcli/cmd/satellite"
	"github.com/infostellarinc/stellarcli/cmd/util"
)

var (
	stellarUse  = util.Normalize("stellar")
	stellarLong = util.Normalize(`stellar is a command line tool for using the StellarStation API.

		To begin, it is generally needed to authenticate the tool by running

		$ stellar auth activate-api-key path/to/stellarstation-private-key.json

		All commands should work after that.`)
	stellarShort = util.Normalize("stellar is a command line tool for using the StellarStation API.")
)

// rootCmd represents the base command when called without any subcommands
func NewRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   stellarUse,
		Short: stellarShort,
		Long:  stellarLong,
	}

	// Add sub commands
	command.AddCommand(auth.NewAuthCommand())
	command.AddCommand(groundstation.GroundStationCmd)
	command.AddCommand(satellite.SatelliteCmd)

	return command
}
