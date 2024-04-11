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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

var listPlansVerboseTemplate = []printer.TemplateItem{
	{Label: "PLAN_ID", Path: "planId"},
	{Label: "AOS_TIME", Path: "aos"},
	{Label: "LOS_TIME", Path: "los"},
	{Label: "SATELLITE_ID", Path: "satelliteInfo.id"},
	{Label: "SATELLITE_ORG_NAME", Path: "satelliteInfo.orgName"},
	{Label: "DL_FREQ_HZ", Path: "downlinkFreq"},
	{Label: "UL_FREQ_HZ", Path: "uplinkFreq"},
	{Label: "TLE_LINE__1", Path: "tleLine1"},
	{Label: "TLE_LINE__2", Path: "tleLine2"},
	{Label: "UNIT_PRICE", Path: "unitPrice"},
}

var listPlansTemplate = []printer.TemplateItem{
	{Label: "PLAN_ID", Path: "planId"},
	{Label: "AOS_TIME", Path: "aos"},
	{Label: "LOS_TIME", Path: "los"},
	{Label: "SATELLITE_ORG_NAME", Path: "satelliteInfo.orgName"},
	{Label: "DL_FREQ_HZ", Path: "downlinkFreq"},
	{Label: "UL_FREQ_HZ", Path: "uplinkFreq"},
	{Label: "UNIT_PRICE", Path: "unitPrice"},
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

	aosAfterTimestamp := timestamppb.New(o.AOSAfter.UTC())

	aosBeforeTimestamp := timestamppb.New(o.AOSBefore.UTC())

	client := groundstation.NewGroundStationServiceClient(conn)
	request := &groundstation.ListPlansRequest{
		GroundStationId: o.ID,
		AosAfter:        aosAfterTimestamp,
		AosBefore:       aosBeforeTimestamp,
	}

	result, err := client.ListPlans(context.Background(), request)
	if err != nil {
		log.Printf("could not list plans: %v\n", err)
		return
	}

	targetTemplate := listPlansTemplate
	if o.IsVerbose {
		targetTemplate = listPlansVerboseTemplate
	}

	defer o.Printer.Flush()
	o.Printer.WriteHeader(targetTemplate)

	var results []map[string]interface{}
	for _, plan := range result.Plan {
		aos := plan.GetAosTime().AsTime()
		los := plan.GetLosTime().AsTime()

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
				"id":      plan.SatelliteId,
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
