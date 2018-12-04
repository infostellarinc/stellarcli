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

package uw

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/infostellarinc/stellarcli/util/printer"
)

type ListUWOptions struct {
	Printer   printer.Printer
	ID        string
	StartTime time.Time
	EndTime   time.Time
}

// Headers of columns
var headers = []interface{}{
	"WINDOW_ID",
	"START_TIME",
	"END_TIME",
}

// ListUW returns a list of unavailability windows for a given ground station.
func ListUW(o *ListUWOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	startTimeTimestamp, err := ptypes.TimestampProto(o.StartTime.UTC())
	if err != nil {
		log.Fatal(err)
	}

	endTimeTimestamp, err := ptypes.TimestampProto(o.EndTime.UTC())
	if err != nil {
		log.Fatal(err)
	}

	client := groundstation.NewGroundStationServiceClient(conn)
	request := &groundstation.ListUnavailabilityWindowsRequest{
		GroundStationId: o.ID,
		StartTime:       startTimeTimestamp,
		EndTime:         endTimeTimestamp,
	}

	result, err := client.ListUnavailabilityWindows(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	defer o.Printer.Flush()
	o.Printer.Write(headers)

	for _, window := range result.Window {
		startTime, err := ptypes.Timestamp(window.StartTime)
		if err != nil {
			log.Fatal(err)
		}

		endTime, err := ptypes.Timestamp(window.EndTime)
		if err != nil {
			log.Fatal(err)
		}

		record := []interface{}{
			window.WindowId,
			startTime,
			endTime,
		}

		o.Printer.Write(record)
	}
}
