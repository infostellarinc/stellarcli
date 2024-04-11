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

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

// Headers of columns
var headers = []interface{}{
	"TLE_LINE_1",
	"TLE_LINE_2",
}

type GetTLEOptions struct {
	Printer     printer.Printer
	SatelliteId string
}

// GetTLE returns a TLE for the given satellite.
func GetTLE(o *GetTLEOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := stellarstation.NewStellarStationServiceClient(conn)
	request := &stellarstation.GetTleRequest{SatelliteId: o.SatelliteId}

	result, err := client.GetTle(context.Background(), request)
	if err != nil {
		log.Printf("problem getting TLE: %v\n", err)
		return
	}

	defer o.Printer.Flush()
	o.Printer.Write(headers)

	record := []interface{}{
		result.Tle.Line_1,
		result.Tle.Line_2,
	}
	o.Printer.Write(record)
}
