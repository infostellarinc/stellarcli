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
	interactiveShort = util.Normalize("Interactive Terminal UI.")
	interactiveLong  = util.Normalize("Interactive Terminal UI.")
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
			log.Printf("Fetching up coming and recent plans for Satellite '%s'.", args[0])
			conn, err := apiclient.Dial()
			if err != nil {
				return fmt.Errorf("problem dialing: %v\n", err)
			}

			client := stellarstation.NewStellarStationServiceClient(conn)

			plansResponse, err := client.ListPlans(cmd.Context(), &stellarstation.ListPlansRequest{
				SatelliteId: args[0],
				AosAfter:    timestamppb.New(time.Now().Add(-60 * time.Minute)),
				AosBefore:   timestamppb.New(time.Now().Add(30 * time.Minute)),
			})
			if err != nil {
				return fmt.Errorf("could not retrieve plans for Satellite '%s': %w", args[0], err)
			}

			var selectedPlan *stellarstation.Plan

			for _, plan := range plansResponse.Plan {
				if plan.Status == stellarstation.Plan_CANCELED {
					continue
				}
				if time.Now().After(plan.GetLosTime().AsTime()) {
					log.Printf("Recent plan '%s' skipped (After LOS).\n", plan.GetId())
					continue
				}
				if selectedPlan == nil {
					selectedPlan = plan
					log.Printf("Plan '%s' selected.\n", plan.GetId())
					break
				}
			}

			if selectedPlan == nil {
				log.Printf("No active or upcoming plans for Satellite '%s'\n", args[0])
				return nil
			}

			model := initialModel(cmd.Context(), client, selectedPlan)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("problem running: %w", err)
			}
			return nil
		},
	}

	return command

}
