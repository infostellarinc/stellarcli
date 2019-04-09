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
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/infostellarinc/stellarcli/util/printer"
)

var listPassesVerboseTemplate = []printer.TemplateItem{
	{"RESERVATION_TOKEN", "reservationToken"},
	{"AOS_TIME", "aos"},
	{"LOS_TIME", "los"},
	{"GS_LAT", "gsInfo.gsLat"},
	{"GS_LONG", "gsInfo.gsLong"},
	{"GS_COUNTRY", "gsInfo.gsCountry"},
	{"MAX_ELEVATION_DEGREE", "maxElevationDegree"},
	{"MAX_ELEVATION_TIME", "maxElevationTime"},
	{"DOWNLINK_CENTER_FREQUENCY_HZ", "downlinkFrequency"},
	{"UPLINK_CENTER_FREQUENCY_HZ", "uplinkFrequency"},
}

var listPassesTemplate = []printer.TemplateItem{
	{"AOS_TIME", "aos"},
	{"LOS_TIME", "los"},
	{"GS_LAT", "gsInfo.gsLat"},
	{"GS_LONG", "gsInfo.gsLong"},
	{"GS_COUNTRY", "gsInfo.gsCountry"},
	{"MAX_ELEVATION_DEGREE", "maxElevationDegree"},
	{"DOWNLINK_CENTER_FREQUENCY_HZ", "downlinkFrequency"},
	{"UPLINK_CENTER_FREQUENCY_HZ", "uplinkFrequency"},
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

	targetTemplate := listPassesTemplate
	if o.IsVerbose {
		targetTemplate = listPassesVerboseTemplate
	}

	defer o.Printer.Flush()
	o.Printer.WriteHeader(targetTemplate)

	var results []map[string]interface{}
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
			obj := map[string]interface{}{
				"reservationToken": pass.ReservationToken,
				"aos":              aos,
				"los":              los,
				"gsInfo": map[string]interface{}{
					"gsLat":     pass.GroundStationLatitude,
					"gsLong":    pass.GroundStationLongitude,
					"gsCountry": pass.GroundStationCountryCode,
				},
				"maxElevationDegree": pass.MaxElevationDegrees,
				"maxElevationTime":   maxElevationTime,
				"downlinkFrequency":  pass.DownlinkCenterFrequencyHz,
				"uplinkFrequency":    pass.UplinkCenterFrequencyHz,
			}

			results = append(results, obj)
		}
	}
	o.Printer.WriteWithTemplate(results, targetTemplate)
}
