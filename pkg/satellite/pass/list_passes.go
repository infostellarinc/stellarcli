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

package pass

import (
	"context"
	"log"

	"github.com/golang/protobuf/ptypes"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/infostellarinc/stellarcli/util/printer"
)

// Headers of columns
var headers = []interface{}{
	"AOS_TIME",
	"LOS_TIME",
	"GS_LAT",
	"GS_LONG",
	"GS_COUNTRY",
	"MAX_ELEVATION_DEGREE",
	"DOWNLINK_CENTER_FREQUENCY_HZ",
	"UPLINK_CENTER_FREQUENCY_HZ",
}

var verboseHeaders = []interface{}{
	"RESERVATION_TOKEN",
	"AOS_TIME",
	"LOS_TIME",
	"GS_LAT",
	"GS_LONG",
	"GS_COUNTRY",
	"MAX_ELEVATION_DEGREE",
	"MAX_ELEVATION_TIME",
	"DOWNLINK_CENTER_FREQUENCY_HZ",
	"UPLINK_CENTER_FREQUENCY_HZ",
}

type ListAvailablePassesOptions struct {
	Printer      printer.Printer
	ID           string
	MinElevation float64
	IsVerbose    bool
}

// ListAvailablePasses returns a list of passes available for a given satellite.
func ListAvailablePasses(o *ListAvailablePassesOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := stellarstation.NewStellarStationServiceClient(conn)
	request := &stellarstation.ListUpcomingAvailablePassesRequest{SatelliteId: o.ID}

	result, err := client.ListUpcomingAvailablePasses(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	targetHeaders := headers
	if o.IsVerbose {
		targetHeaders = verboseHeaders
	}

	defer o.Printer.Flush()
	o.Printer.Write(targetHeaders)

	for _, pass := range result.Pass {
		aos, err := ptypes.Timestamp(pass.AosTime)
		if err != nil {
			log.Fatal(err)
		}

		los, err := ptypes.Timestamp(pass.LosTime)
		if err != nil {
			log.Fatal(err)
		}

		maxElevationTime, err := ptypes.Timestamp(pass.MaxElevationTime)
		if err != nil {
			log.Fatal(err)
		}

		if pass.MaxElevationDegrees > o.MinElevation {
			values := map[interface{}]interface{}{
				"RESERVATION_TOKEN":            pass.ReservationToken,
				"AOS_TIME":                     aos,
				"LOS_TIME":                     los,
				"GS_LAT":                       pass.GroundStationLatitude,
				"GS_LONG":                      pass.GroundStationLongitude,
				"GS_COUNTRY":                   pass.GroundStationCountryCode,
				"MAX_ELEVATION_DEGREE":         pass.MaxElevationDegrees,
				"MAX_ELEVATION_TIME":           maxElevationTime,
				"DOWNLINK_CENTER_FREQUENCY_HZ": pass.DownlinkCenterFrequencyHz,
				"UPLINK_CENTER_FREQUENCY_HZ":   pass.UplinkCenterFrequencyHz,
			}

			o.Printer.Write(util.MapToSlice(values, targetHeaders))

		}
	}
}
