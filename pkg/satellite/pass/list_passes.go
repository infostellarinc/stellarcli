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

package pass

import (
	"context"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

var listPassesVerboseTemplate = []printer.TemplateItem{
	{Label: "RESERVATION_TOKEN", Path: "reservationToken"},
	{Label: "AOS_TIME", Path: "aos"},
	{Label: "LOS_TIME", Path: "los"},
	{Label: "GS_ID", Path: "gsInfo.id"},
	{Label: "GS_ORG_NAME", Path: "gsInfo.orgName"},
	{Label: "GS_LAT", Path: "gsInfo.latitude"},
	{Label: "GS_LONG", Path: "gsInfo.longitude"},
	{Label: "GS_COUNTRY", Path: "gsInfo.country"},
	{Label: "MAX_ELEVATION_DEGREE", Path: "maxElevationDegree"},
	{Label: "MAX_ELEVATION_TIME", Path: "maxElevationTime"},
	{Label: "CHANNEL_SET_ID", Path: "channelSet.id"},
	{Label: "CHANNEL_SET_NAME", Path: "channelSet.name"},
	{Label: "DL_FREQ_HZ", Path: "channelSet.downlink.frequency"},
	{Label: "DL_MODULATION", Path: "channelSet.downlink.modulation"},
	{Label: "DL_BITRATE", Path: "channelSet.downlink.bitrate"},
	{Label: "UL_FREQ_HZ", Path: "channelSet.uplink.frequency"},
	{Label: "UL_MODULATION", Path: "channelSet.uplink.modulation"},
	{Label: "UL_BITRATE", Path: "channelSet.uplink.bitrate"},
	{Label: "UNIT_PRICE", Path: "unitPrice"},
}

var listPassesTemplate = []printer.TemplateItem{
	{Label: "AOS_TIME", Path: "aos"},
	{Label: "LOS_TIME", Path: "los"},
	{Label: "GS_LAT", Path: "gsInfo.latitude"},
	{Label: "GS_LONG", Path: "gsInfo.longitude"},
	{Label: "GS_COUNTRY", Path: "gsInfo.country"},
	{Label: "MAX_ELEVATION_DEGREE", Path: "maxElevationDegree"},
	{Label: "CHANNEL_SET_NAME", Path: "channelSet.name"},
	{Label: "DL_FREQ_HZ", Path: "channelSet.downlink.frequency"},
	{Label: "UL_FREQ_HZ", Path: "channelSet.uplink.frequency"},
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
		log.Printf("problem fetching upcoming passes: %v", err)
		return
	}

	targetTemplate := listPassesTemplate
	if o.IsVerbose {
		targetTemplate = listPassesVerboseTemplate
	}

	defer o.Printer.Flush()
	o.Printer.WriteHeader(targetTemplate)

	var results []map[string]interface{}
	for _, pass := range result.Pass {
		aos := pass.GetAosTime().AsTime()
		los := pass.GetLosTime().AsTime()
		maxElevationTime := pass.GetMaxElevationTime().AsTime()

		if pass.MaxElevationDegrees > o.MinElevation {
			channelSetTokens := pass.ChannelSetToken
			for _, channelSetToken := range channelSetTokens {
				channelSet := channelSetToken.ChannelSet

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
					"reservationToken": channelSetToken.ReservationToken,
					"aos":              aos,
					"los":              los,
					"gsInfo": map[string]interface{}{
						"id":        pass.GroundStationId,
						"latitude":  pass.GroundStationLatitude,
						"longitude": pass.GroundStationLongitude,
						"country":   pass.GroundStationCountryCode,
						"orgName":   pass.GroundStationOrganizationName,
					},
					"maxElevationDegree": pass.MaxElevationDegrees,
					"maxElevationTime":   maxElevationTime,
					"unitPrice":          channelSetToken.UnitPrice,
					"channelSet": map[string]interface{}{
						"id":       channelSet.Id,
						"name":     channelSet.Name,
						"downlink": downlink,
						"uplink":   uplink,
					},
				}

				results = append(results, obj)
			}
		}
	}
	o.Printer.WriteWithTemplate(results, targetTemplate)
}
