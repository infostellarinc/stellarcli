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
	{"GS_ID", "gsInfo.id"},
	{"GS_ORG_NAME", "gsInfo.orgName"},
	{"GS_LAT", "gsInfo.latitude"},
	{"GS_LONG", "gsInfo.longitude"},
	{"GS_COUNTRY", "gsInfo.country"},
	{"MAX_ELEVATION_DEGREE", "maxElevationDegree"},
	{"MAX_ELEVATION_TIME", "maxElevationTime"},
	{"CHANNEL_SET_ID", "channelSet.id"},
	{"CHANNEL_SET_NAME", "channelSet.name"},
	{"DL_FREQ_HZ", "channelSet.downlink.frequency"},
	{"DL_MODULATION", "channelSet.downlink.modulation"},
	{"DL_BITRATE", "channelSet.downlink.bitrate"},
	{"UL_FREQ_HZ", "channelSet.uplink.frequency"},
	{"UL_MODULATION", "channelSet.uplink.modulation"},
	{"UL_BITRATE", "channelSet.uplink.bitrate"},
	{"UNIT_PRICE", "unitPrice"},
}

var listPassesTemplate = []printer.TemplateItem{
	{"AOS_TIME", "aos"},
	{"LOS_TIME", "los"},
	{"GS_LAT", "gsInfo.latitude"},
	{"GS_LONG", "gsInfo.longitude"},
	{"GS_COUNTRY", "gsInfo.country"},
	{"MAX_ELEVATION_DEGREE", "maxElevationDegree"},
	{"CHANNEL_SET_NAME", "channelSet.name"},
	{"DL_FREQ_HZ", "channelSet.downlink.frequency"},
	{"UL_FREQ_HZ", "channelSet.uplink.frequency"},
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
