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
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

type AddUWOptions struct {
	Printer   printer.Printer
	ID        string
	StartTime time.Time
	EndTime   time.Time
}

// AddUW adds a new unavailability uw to a given ground station.
func AddUW(o *AddUWOptions) {
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
	request := &groundstation.AddUnavailabilityWindowRequest{
		GroundStationId: o.ID,
		StartTime:       startTimeTimestamp,
		EndTime:         endTimeTimestamp,
	}

	result, err := client.AddUnavailabilityWindow(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	defer o.Printer.Flush()
	message := fmt.Sprintf("Succeeded to add the unavailability window as: %s", result.WindowId)
	o.Printer.Write([]interface{}{message})
}
