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
	"io/ioutil"
	"os"
	"path/filepath"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"

	"github.com/infostellarinc/stellarcli/pkg/config"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
)

// NewDefaultCredentials initializes gRPC credentials using Stellar Default Credentials.
func NewDefaultCredentials() (credentials.PerRPCCredentials, error) {
	return oauth.NewJWTAccessFromFile(findDefaultCredentials())
}

// StoreCredentialsFile stores the API key at the given path to a well known location.
func StoreCredentialsFile(path string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("could not read credentials file.\n%v", err)
	}

	config.EnsureConfigDir()

	err = ioutil.WriteFile(wellKnownFile(), content, 0600)
	if err != nil {
		log.Fatalf("could not write to config directory.\n%v", err)
	}
}

func findDefaultCredentials() string {
	// First, try the environment variable.
	const envVar = "STELLAR_CREDENTIALS"
	if filename := os.Getenv(envVar); filename != "" {
		return filename
	}

	// Second, try a well-known file.
	return wellKnownFile()
}

func wellKnownFile() string {
	const f = "stellarstation_credentials.json"
	return filepath.Join(config.GetConfigDir(), f)
}
