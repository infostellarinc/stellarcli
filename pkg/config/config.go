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
	"os"
	"path/filepath"

	log "github.com/infostellarinc/stellarcli/pkg/logger"
)

// EnsureConfigDir ensures the configuration directory exists, creating it and all parents as required.
func EnsureConfigDir() error {
	return os.MkdirAll(GetConfigDir(), 0755)
}

// GetConfigDir returns the directory containing configuration files for stellar.
func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("could not find config directory.\n%v", err)
	}

	return filepath.Join(configDir, "stellar")
}
