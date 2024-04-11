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

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	log "github.com/infostellarinc/stellarcli/pkg/logger"
	"github.com/infostellarinc/stellarcli/util/printer"
)

type DeleteUWOptions struct {
	Printer  printer.Printer
	WindowID string
}

// DeleteUW deletes the existing unavailability window.
func DeleteUW(o *DeleteUWOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := groundstation.NewGroundStationServiceClient(conn)
	request := &groundstation.DeleteUnavailabilityWindowRequest{
		WindowId: o.WindowID,
	}

	_, err = client.DeleteUnavailabilityWindow(context.Background(), request)
	if err != nil {
		log.Printf("problem deleting unavailability window: %v\n", err)
		return
	}

	defer o.Printer.Flush()
	message := fmt.Sprintf("Succeeded to delete the unavailability window: %s", o.WindowID)
	o.Printer.Write([]interface{}{message})
}
