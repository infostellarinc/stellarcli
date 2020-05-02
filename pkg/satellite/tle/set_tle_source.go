// Copyright © 2019 Infostellar, Inc.
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

package tle

import (
	"context"
	"fmt"
	"strings"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

type SetTLESourceOptions struct {
	Printer     printer.Printer
	SatelliteID string
	Source      string
}

// SetTleSource set the TLE source for a given satellite.
func SetTLESource(o *SetTLESourceOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := stellarstation.NewStellarStationServiceClient(conn)

	var sourceOption = strings.ToLower(o.Source)
	var source stellarstation.SetTleSourceRequest_Source

	if sourceOption == "manual" {
		source = stellarstation.SetTleSourceRequest_MANUAL
	} else if sourceOption == "norad" {
		source = stellarstation.SetTleSourceRequest_NORAD
	} else {
		log.Fatal("Invalid source provided")
	}

	request := &stellarstation.SetTleSourceRequest{
		SatelliteId: o.SatelliteID,
		Source:      source,
	}

	_, err = client.SetTleSource(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	defer o.Printer.Flush()
	message := fmt.Sprintf("Successfully changed TLE source.")
	o.Printer.Write([]interface{}{message})
}
