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

var listPlansVerboseTemplate = []printer.TemplateItem{
	{"PLAN_ID", "planId"},
	{"AOS_TIME", "aos"},
	{"LOS_TIME", "los"},
	{"SATELLITE_ORG_NAME", "satelliteInfo.orgName"},
	{"DL_FREQ_HZ", "downlinkFreq"},
	{"UL_FREQ_HZ", "uplinkFreq"},
	{"TLE_LINE__1", "tleLine1"},
	{"TLE_LINE__2", "tleLine2"},
	{"UNIT_PRICE", "unitPrice"},
}

var listPlansTemplate = []printer.TemplateItem{
	{"PLAN_ID", "planId"},
	{"AOS_TIME", "aos"},
	{"LOS_TIME", "los"},
	{"SATELLITE_ORG_NAME", "satelliteInfo.orgName"},
	{"DL_FREQ_HZ", "downlinkFreq"},
	{"UL_FREQ_HZ", "uplinkFreq"},
	{"UNIT_PRICE", "unitPrice"},
}

type ListOptions struct {
	Printer   printer.Printer
	ID        string
	AOSAfter  *time.Time
	AOSBefore *time.Time
	IsVerbose bool
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

	targetTemplate := listPlansTemplate
	if o.IsVerbose {
		targetTemplate = listPlansVerboseTemplate
	}

	defer o.Printer.Flush()
	o.Printer.WriteHeader(targetTemplate)

	var results []map[string]interface{}
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
		obj := map[string]interface{}{
			"planId": plan.PlanId,
			"aos":    aos,
			"los":    los,
			"satelliteInfo": map[string]interface{}{
				"orgName": plan.SatelliteOrganizationName,
			},
			"unitPrice":    plan.UnitPrice,
			"downlinkFreq": downlinkFreq,
			"uplinkFreq":   uplinkFreq,
			"tleLine1":     plan.Tle.Line_1,
			"tleLine2":     plan.Tle.Line_2,
		}
		results = append(results, obj)
	}
	o.Printer.WriteWithTemplate(results, targetTemplate)
}
