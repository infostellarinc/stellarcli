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

package apiclient

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/infostellarinc/stellarcli/cmd/app"
	"github.com/infostellarinc/stellarcli/pkg/auth"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
)

// Dial opens a gRPC connection to the StellarStation API with authentication setup.
func Dial() (*grpc.ClientConn, error) {
	creds, err := auth.NewDefaultCredentials()
	if err != nil {
		return nil, err
	}

	apiUrl := os.Getenv("STELLARSTATION_API_URL")
	if len(apiUrl) == 0 {
		apiUrl = "api.stellarstation.com:443"
	}
	log.Printf("API endpoint: %s", apiUrl)

	tlsConfig := &tls.Config{}
	if strings.HasPrefix(apiUrl, "localhost") || strings.HasPrefix(apiUrl, "127.0.0.1") {
		tlsConfig.InsecureSkipVerify = true
	}

	return grpc.Dial(
		apiUrl,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(creds),
		grpc.WithUserAgent(fmt.Sprintf("stellarcli/%s", app.ReleaseVersion)),
		// Set receive size to a somewhat safe 9MiB.
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(9437000)))
}
