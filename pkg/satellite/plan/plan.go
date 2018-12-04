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

package plan

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/infostellarinc/stellarcli/util/printer"
)

// Headers of columns
var headers = []interface{}{
	"ID",
	"SATELLITE_ID",
	"STATUS",
	"AOS_TIME",
	"LOS_TIME",
	"GS_LAT",
	"GS_LONG",
	"GS_COUNTRY",
	"MAX_ELEVATION_DEGREE",
	"MAX_ELEVATION_TIME",
	"DL_FREQ_HZ",
	"UL_FREQ_HZ",
}

type ListOptions struct {
	Printer   printer.Printer
	ID        string
	AOSAfter  *time.Time
	AOSBefore *time.Time
}

// ListPlans returns a list of plans for a given ground station.
func ListPlans(o *ListOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	aosAfterTimestamp, err := ptypes.TimestampProto(o.AOSAfter.UTC())
	if err != nil {
		log.Fatal(err)
	}

	aosBeforeTimestamp, err := ptypes.TimestampProto(o.AOSBefore.UTC())
	if err != nil {
		log.Fatal(err)
	}

	client := stellarstation.NewStellarStationServiceClient(conn)
	request := &stellarstation.ListPlansRequest{
		SatelliteId: o.ID,
		AosAfter:    aosAfterTimestamp,
		AosBefore:   aosBeforeTimestamp,
	}

	result, err := client.ListPlans(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	defer o.Printer.Flush()
	o.Printer.Write(headers)

	for _, plan := range result.Plan {
		aos, err := ptypes.Timestamp(plan.AosTime)
		if err != nil {
			log.Fatal(err)
		}

		los, err := ptypes.Timestamp(plan.LosTime)
		if err != nil {
			log.Fatal(err)
		}

		maxElevationTime, err := ptypes.Timestamp(plan.MaxElevationTime)
		if err != nil {
			log.Fatal(err)
		}

		record := []interface{}{
			plan.Id,
			plan.SatelliteId,
			plan.Status,
			aos,
			los,
			plan.GroundStationLatitude,
			plan.GroundStationLongitude,
			plan.GroundStationCountryCode,
			plan.MaxElevationDegrees,
			maxElevationTime,
			plan.DownlinkCenterFrequencyHz,
			plan.UplinkCenterFrequencyHz,
		}
		o.Printer.Write(record)
	}
}
