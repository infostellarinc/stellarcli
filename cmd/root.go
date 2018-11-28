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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infostellarinc/stellarcli/cmd/auth"
	"github.com/infostellarinc/stellarcli/cmd/groundstation"
	"github.com/infostellarinc/stellarcli/cmd/satellite"
	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/config"
)

var (
	stellarUse  = util.Normalize("stellar")
	stellarLong = util.Normalize(`stellar is a command line tool for using the StellarStation API.

		To begin, it is generally needed to authenticate the tool by running

		$ stellar auth activate-api-key path/to/stellarstation-private-key.json

		All commands should work after that.`)
	stellarShort = util.Normalize("stellar is a command line tool for using the StellarStation API.")
)
var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   stellarUse,
	Short: stellarShort,
	Long:  stellarLong,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/config/stellar/config.yaml)")

	// Add sub commands
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(groundstation.GroundStationCmd)
	rootCmd.AddCommand(satellite.SatelliteCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(config.GetConfigDir())
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
