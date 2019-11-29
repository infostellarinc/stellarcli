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
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/infostellarinc/stellarcli/util/printer"
)

var listPlansTemplate = []printer.TemplateItem{
	{"ID", "id"},
	{"SATELLITE_ID", "satelliteId"},
	{"CHANNEL_SET_NAME", "channelSet.name"},
	{"STATUS", "status"},
	{"AOS_TIME", "aos"},
	{"LOS_TIME", "los"},
	{"GS_ORG_NAME", "gsInfo.orgName"},
	{"GS_COUNTRY", "gsInfo.country"},
	{"MAX_ELEVATION_DEGREE", "maxElevationDegree"},
	{"UNIT_PRICE", "unitPrice"},
}

var listPlansVerboseTemplate = []printer.TemplateItem{
	{"ID", "id"},
	{"SATELLITE_ID", "satelliteId"},
	{"CHANNEL_SET_NAME", "channelSet.name"},
	{"CHANNEL_SET_ID", "channelSet.id"},
	{"STATUS", "status"},
	{"AOS_TIME", "aos"},
	{"LOS_TIME", "los"},
	{"GS_ORG_NAME", "gsInfo.orgName"},
	{"GS_LAT", "gsInfo.latitude"},
	{"GS_LONG", "gsInfo.longitude"},
	{"GS_COUNTRY", "gsInfo.country"},
	{"MAX_ELEVATION_DEGREE", "maxElevationDegree"},
	{"MAX_ELEVATION_TIME", "maxElevationTime"},
	{"DL_FREQ_HZ", "channelSet.downlink.frequency"},
	{"UL_FREQ_HZ", "channelSet.uplink.frequency"},
	{"UNIT_PRICE", "unitPrice"},
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

	targetTemplate := listPlansTemplate

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

		maxElevationTime, err := ptypes.Timestamp(plan.MaxElevationTime)
		if err != nil {
			log.Fatal(err)
		}

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
				"latitude":  plan.GroundStationLatitude,
				"longitude": plan.GroundStationLongitude,
				"country":   plan.GroundStationCountryCode,
				"orgName":   plan.GroundStationOrganizationName,
			},
			"unitPrice":          plan.UnitPrice,
			"maxElevationDegree": plan.MaxElevationDegrees,
			"maxElevationTime":   maxElevationTime,
		}
		results = append(results, obj)
	}
	o.Printer.WriteWithTemplate(results, targetTemplate)
}
