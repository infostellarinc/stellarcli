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
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/infostellarinc/stellarcli/pkg/logger"
)

const (
	SkipSSLVerify = true
)

// Dial opens a gRPC connection to the StellarStation API with authentication setup.
func Dial() (*grpc.ClientConn, error) {
	apiUrl := os.Getenv("STELLARSTATION_API_URL")
	if len(apiUrl) == 0 {
		apiUrl = "api.stellarstation.com:443"
	}
	log.Printf("API endpoint: %s", apiUrl)

	certPath := os.Getenv("STELLARCLI_TLS_CERT_PATH")
	keyPath := os.Getenv("STELLARCLI_TLS_KEY_PATH")
	caPath := os.Getenv("STELLARCLI_TLS_CA_PATH")

	tls, err := NewTLSCreds(certPath, keyPath, caPath)
	if err != nil {
		log.Printf("failed to setup tls")
	}

	return grpc.Dial(
		apiUrl,
		grpc.WithTransportCredentials(tls),
		// Set receive size to a somewhat safe 9MiB.
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(9437000)))
}

func NewTLSCreds(certPath, keyPath, caPath string) (credentials.TransportCredentials, error) {
	if ok, err := exists(certPath); err != nil || !ok {
		return nil, fmt.Errorf("missing file mtls certificate")
	}

	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		RootCAs:            certPool,
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: SkipSSLVerify,
	}), nil
}

func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	return false, fmt.Errorf("could not load file (%v): %w", name, err)
}
