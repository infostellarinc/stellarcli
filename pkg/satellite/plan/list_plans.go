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

package plan

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

var listPlansVerboseTemplate = []printer.TemplateItem{
	{Label: "ID", Path: "id"},
	{Label: "SATELLITE_ID", Path: "satelliteId"},
	{Label: "CHANNEL_SET_NAME", Path: "channelSet.name"},
	{Label: "CHANNEL_SET_ID", Path: "channelSet.id"},
	{Label: "STATUS", Path: "status"},
	{Label: "AOS_TIME", Path: "aos"},
	{Label: "LOS_TIME", Path: "los"},
	{Label: "GS_ID", Path: "gsInfo.id"},
	{Label: "GS_ORG_NAME", Path: "gsInfo.orgName"},
	{Label: "GS_LAT", Path: "gsInfo.latitude"},
	{Label: "GS_LONG", Path: "gsInfo.longitude"},
	{Label: "GS_COUNTRY", Path: "gsInfo.country"},
	{Label: "MAX_ELEVATION_DEGREE", Path: "maxElevationDegree"},
	{Label: "MAX_ELEVATION_TIME", Path: "maxElevationTime"},
	{Label: "DL_FREQ_HZ", Path: "channelSet.downlink.frequency"},
	{Label: "UL_FREQ_HZ", Path: "channelSet.uplink.frequency"},
	{Label: "UNIT_PRICE", Path: "unitPrice"},
}

var listPlansTemplate = []printer.TemplateItem{
	{Label: "ID", Path: "id"},
	{Label: "SATELLITE_ID", Path: "satelliteId"},
	{Label: "CHANNEL_SET_NAME", Path: "channelSet.name"},
	{Label: "STATUS", Path: "status"},
	{Label: "AOS_TIME", Path: "aos"},
	{Label: "LOS_TIME", Path: "los"},
	{Label: "GS_ORG_NAME", Path: "gsInfo.orgName"},
	{Label: "GS_COUNTRY", Path: "gsInfo.country"},
	{Label: "MAX_ELEVATION_DEGREE", Path: "maxElevationDegree"},
	{Label: "UNIT_PRICE", Path: "unitPrice"},
}

type ListOptions struct {
	Printer   printer.Printer
	ID        string
	AOSAfter  *time.Time
	AOSBefore *time.Time
	IsVerbose bool
}

// ListPlans returns a list of plans for a given satellite.
func ListPlans(o *ListOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	aosAfterTimestamp := timestamppb.New(o.AOSAfter.UTC())
	aosBeforeTimestamp := timestamppb.New(o.AOSBefore.UTC())

	client := stellarstation.NewStellarStationServiceClient(conn)
	request := &stellarstation.ListPlansRequest{
		SatelliteId: o.ID,
		AosAfter:    aosAfterTimestamp,
		AosBefore:   aosBeforeTimestamp,
	}

	result, err := client.ListPlans(context.Background(), request)
	if err != nil {
		log.Printf("error listing plans: %v", err)
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
		maxElevationTime := plan.GetMaxElevationTime().AsTime()

		channelSet := plan.ChannelSet
		downlink := make(map[string]interface{})
		if channelSet.Downlink != nil {
			downlink["frequency"] = channelSet.Downlink.CenterFrequencyHz
			downlink["modulation"] = channelSet.Downlink.Modulation
			downlink["protocol"] = channelSet.Downlink.Protocol
			downlink["bitrate"] = channelSet.Downlink.Bitrate
		}
		uplink := make(map[string]interface{})
		if channelSet.Uplink != nil {
			uplink["frequency"] = channelSet.Uplink.CenterFrequencyHz
			uplink["modulation"] = channelSet.Uplink.Modulation
			uplink["protocol"] = channelSet.Uplink.Protocol
			uplink["bitrate"] = channelSet.Uplink.Bitrate
		}

		var telemetryMetadata []map[string]interface{}
		if len(plan.TelemetryMetadata) > 0 {
			for _, metadata := range plan.TelemetryMetadata {
				data := make(map[string]interface{})
				data["url"] = metadata.Url
				data["dataType"] = metadata.DataType
				telemetryMetadata = append(telemetryMetadata, data)
			}
		}

		obj := map[string]interface{}{
			"id":          plan.Id,
			"satelliteId": plan.SatelliteId,
			"channelSet": map[string]interface{}{
				"id":       plan.ChannelSet.Id,
				"name":     plan.ChannelSet.Name,
				"downlink": downlink,
				"uplink":   uplink,
			},
			"status": plan.Status,
			"aos":    aos,
			"los":    los,
			"gsInfo": map[string]interface{}{
				"id":        plan.GroundStationId,
				"latitude":  plan.GroundStationLatitude,
				"longitude": plan.GroundStationLongitude,
				"country":   plan.GroundStationCountryCode,
				"orgName":   plan.GroundStationOrganizationName,
			},
			"unitPrice":          plan.UnitPrice,
			"maxElevationDegree": plan.MaxElevationDegrees,
			"maxElevationTime":   maxElevationTime,
			"telemetryMetadata":  telemetryMetadata,
		}
		results = append(results, obj)
	}
	o.Printer.WriteWithTemplate(results, targetTemplate)
}
