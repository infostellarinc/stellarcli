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
	"fmt"
	"log"

	"github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/infostellarinc/stellarcli/util/printer"
)

type CancelPlanOptions struct {
	Printer printer.Printer
	PlanID  string
}

// CancelPlan cancels a plan.
func CancelPlan(o *CancelPlanOptions) {
	conn, err := apiclient.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := groundstation.NewGroundStationServiceClient(conn)
	request := &groundstation.CancelPlanRequest{
		PlanId: o.PlanID,
	}

	_, err = client.CancelPlan(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}

	defer o.Printer.Flush()
	message := fmt.Sprintf("Succeeded to cancel the plan: %s", request.PlanId)
	o.Printer.Write([]interface{}{message})
}
