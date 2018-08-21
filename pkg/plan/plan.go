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
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
)

// Default format of time.Timestamp when conerting it to a textual representation
const defaultFormat = time.RFC3339

// Default separator between columns
const defaultSeparator = ","

// Headers of columns
var headers = []string{
	"PLAN_ID",
	"AOS_TIME",
	"LOS_TIME",
	"DOWNLINK_CENTER_FREQUENCY_HZ",
	"UPLINK_CENTER_FREQUENCY_HZ",
	"TLE_LINE__1",
	"TLE_LINE__2",
}

// ListPlans returns a list of plans for a given ground staion
func ListPlans(id string, aosAfter, aosBefore time.Time) {

	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	aosAfterTimestamp, err := ptypes.TimestampProto(aosAfter.UTC())
	if err != nil {
		log.Fatal(err)
	}

	aosBeforeTimestamp, err := ptypes.TimestampProto(aosBefore.UTC())
	if err != nil {
		log.Fatal(err)
	}

	client := groundstation.NewGroundStationServiceClient(conn)
	request := &groundstation.ListPlansRequest{
		GroundStationId: id,
		AosAfter:        aosAfterTimestamp,
		AosBefore:       aosBeforeTimestamp,
	}

	result, err := client.ListPlans(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	// TODO(hoshir): Accepts time format from commandline options
	layout := defaultFormat

	// TODO(hoshir): Accepts a separator from commandline options
	sep := defaultSeparator

	fmt.Println(strings.Join(headers, sep))
	for _, plan := range result.Plan {
		aos, err := ptypes.Timestamp(plan.AosTime)
		if err != nil {
			log.Fatal(err)
		}

		los, err := ptypes.Timestamp(plan.LosTime)
		if err != nil {
			log.Fatal(err)
		}

		var b bytes.Buffer
		b.WriteString(fmt.Sprintf("%q%s", plan.PlanId, sep))
		b.WriteString(fmt.Sprintf("%q%s", aos.Format(layout), sep))
		b.WriteString(fmt.Sprintf("%q%s", los.Format(layout), sep))
		b.WriteString(fmt.Sprintf("%v%s", plan.DownlinkCenterFrequencyHz, sep))
		b.WriteString(fmt.Sprintf("%v%s", plan.UplinkCenterFrequencyHz, sep))
		b.WriteString(fmt.Sprintf("%q%s", plan.Tle.Line_1, sep))
		b.WriteString(fmt.Sprintf("%q%s", plan.Tle.Line_2, sep))

		fmt.Println(b.String())
	}
}
