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

package tle

import (
	"context"
	"fmt"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/go-stellarstation/api/v1/orbit"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

type AddTLEOptions struct {
	Printer     printer.Printer
	SatelliteID string
	Line1       string
	Line2       string
}

// AddTLE adds a new TLE to a given satellite.
func AddTLE(o *AddTLEOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := stellarstation.NewStellarStationServiceClient(conn)
	tle := &orbit.Tle{
		Line_1: o.Line1,
		Line_2: o.Line2,
	}
	request := &stellarstation.AddTleRequest{
		SatelliteId: o.SatelliteID,
		Tle:         tle,
	}

	_, err = client.AddTle(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	defer o.Printer.Flush()
	message := fmt.Sprintf("Succeeded to add the TLE.")
	o.Printer.Write([]interface{}{message})
}
