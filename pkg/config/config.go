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

package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
)

// EnsureConfigDir ensures the configuration directory exists, creating it and all parents as required.
func EnsureConfigDir() {
	os.MkdirAll(GetConfigDir(), 0755)
}

// GetConfigDir returns the directory containing configuration files for stellar.
func GetConfigDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "stellar")
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("could not find home directory.\n%v", err)
	}

	return filepath.Join(home, ".config", "stellar")
}
