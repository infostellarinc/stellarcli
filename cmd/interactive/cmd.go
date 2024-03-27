package interactive

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/pkg/apiclient"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	interactiveUse   = util.Normalize("interactive-plan [Satellite ID]")
	interactiveShort = util.Normalize("Interactive Terminal UI (experimental).")
	interactiveLong  = util.Normalize("Experimental Interactive Terminal UI that supports simple telemetry receiving checks and modem configuration changes.")
)

// Create reserve-pass command.
func NewInteractiveCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   interactiveUse,
		Short: interactiveShort,
		Long:  interactiveLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			debugMode, _ := cmd.PersistentFlags().GetBool("debug")

			log.Printf("Fetching up coming and recent plans for Satellite '%s'.", args[0])
			conn, err := apiclient.Dial()
			if err != nil {
				// don't return an error as it's not a CLI error
				log.Printf("problem dialing: %v\n", err)
				return nil
			}

			client := stellarstation.NewStellarStationServiceClient(conn)

			plansResponse, err := client.ListPlans(cmd.Context(), &stellarstation.ListPlansRequest{
				SatelliteId: args[0],
				AosAfter:    timestamppb.New(time.Now().Add(-60 * time.Minute)),
				AosBefore:   timestamppb.New(time.Now().Add(30 * time.Minute)),
			})
			if err != nil {
				log.Printf("could not retrieve plans for Satellite '%s': %v\n", args[0], err)
				// don't return an error as it's not a CLI error
				return nil
			}

			var selectedPlan *stellarstation.Plan

			for _, plan := range plansResponse.Plan {
				if plan.Status == stellarstation.Plan_CANCELED {
					continue
				}

				endTime := getEndTime(plan)

				if time.Now().After(endTime) {
					log.Printf("Recent plan '%s' skipped (After Operation End).\n", plan.GetId())
					continue
				}
				if selectedPlan == nil {
					selectedPlan = plan
					break
				}
			}

			if selectedPlan == nil {
				log.Printf("No active or upcoming plans for Satellite '%s'\n", args[0])
				printRecentPlans(plansResponse.Plan)
				return nil
			}

			model := initialModel(
				cmd.Context(),
				client,
				selectedPlan,
				debugMode,
			)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("problem running: %w", err)
			}
			return nil
		},
	}

	return command

}

func getEndTime(plan *stellarstation.Plan) time.Time {
	// prefer operation end
	if endTime := plan.GetEndTime(); endTime != nil {
		return endTime.AsTime()
	}
	if los := plan.GetLosTime(); los != nil {
		return los.AsTime()
	}

	return time.Time{}
}

func printRecentPlans(plans []*stellarstation.Plan) {
	log.Println("Recent Plans:")
	for _, plan := range plans {
		log.Printf(" - Plan %v: %v %v - %v\n",
			plan.Id, plan.GetStatus(),
			plan.GetAosTime().AsTime().UTC().Format(time.RFC3339),
			plan.GetLosTime().AsTime().UTC().Format(time.RFC3339),
		)
	}
}
