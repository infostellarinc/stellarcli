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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
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

	startTimeTimestamp := timestamppb.New(o.StartTime.UTC())
	endTimeTimestamp := timestamppb.New(o.EndTime.UTC())

	client := groundstation.NewGroundStationServiceClient(conn)
	request := &groundstation.ListUnavailabilityWindowsRequest{
		GroundStationId: o.ID,
		StartTime:       startTimeTimestamp,
		EndTime:         endTimeTimestamp,
	}

	result, err := client.ListUnavailabilityWindows(context.Background(), request)
	if err != nil {
		log.Printf("problem listing unavailability windows: %v\n", err)
		return
	}

	defer o.Printer.Flush()
	o.Printer.Write(headers)

	for _, window := range result.Window {
		startTime := window.GetStartTime().AsTime()
		endTime := window.GetEndTime().AsTime()

		record := []interface{}{
			window.WindowId,
			startTime,
			endTime,
		}

		o.Printer.Write(record)
	}
}
