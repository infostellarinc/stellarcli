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

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/infostellarinc/stellarcli/util/printer"
)

type ListOptions struct {
	Printer   printer.Printer
	ID        string
	AOSAfter  *time.Time
	AOSBefore *time.Time
}

// Headers of columns
var headers = []interface{}{
	"PLAN_ID",
	"AOS_TIME",
	"LOS_TIME",
	"DOWNLINK_CENTER_FREQUENCY_HZ",
	"UPLINK_CENTER_FREQUENCY_HZ",
	"TLE_LINE__1",
	"TLE_LINE__2",
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

	client := groundstation.NewGroundStationServiceClient(conn)
	request := &groundstation.ListPlansRequest{
		GroundStationId: o.ID,
		AosAfter:        aosAfterTimestamp,
		AosBefore:       aosBeforeTimestamp,
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

		var downlinkFreq, uplinkFreq uint64
		if plan.DownlinkRadioDevice != nil {
			downlinkFreq = plan.DownlinkRadioDevice.CenterFrequencyHz
		}
		if plan.UplinkRadioDevice != nil {
			uplinkFreq = plan.UplinkRadioDevice.CenterFrequencyHz
		}
		record := []interface{}{
			plan.PlanId,
			aos,
			los,
			downlinkFreq,
			uplinkFreq,
			plan.Tle.Line_1,
			plan.Tle.Line_2,
		}

		o.Printer.Write(record)
	}
}
