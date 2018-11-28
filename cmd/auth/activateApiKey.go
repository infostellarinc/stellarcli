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

package auth

import (
	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/auth"
)

var (
	activateApiKeyUse   = util.Normalize("activate-api-key [path-to-key]")
	activateApiKeyShort = util.Normalize("Activate an API key for use in following commands.")
	activateApiKeyLong  = util.Normalize(`Activates an API key for use in following commands by copying it to the
		configuration directory.`)
)

// activateApiKeyCmd represents the activateApiKey command
var activateApiKeyCmd = &cobra.Command{
	Use:   activateApiKeyUse,
	Short: activateApiKeyShort,
	Long:  activateApiKeyLong,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		auth.StoreCredentialsFile(args[0])
	},
}
