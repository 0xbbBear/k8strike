package cli

import (
	"fmt"
	"log"

	"k8strike/pkg/evaluate"
	"github.com/spf13/cobra"
)

var EvaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "Gather information to find weakness inside container",
	Long: `Evaluate performs security checks against the current container environment.
It gathers information about the container runtime, capabilities, mounted resources,
network configuration, and other security-relevant aspects.`,
	Aliases: []string{"eva"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(BannerHeader)
		profileID := evaluate.ProfileBasic
		if full {
			profileID = evaluate.ProfileExtended
		}
		if profile != "" {
			profileID = profile
		}
		if err := evaluate.NewEvaluator().RunProfile(profileID, nil); err != nil {
			log.Printf("evaluate profile %q failed: %v", profileID, err)
		}
	},
}

var full bool
var profile string

func init() {
	EvaluateCmd.Flags().BoolVarP(&full, "full", "", false, "Enable file scan during information gathering (equivalent to --profile extended)")
	EvaluateCmd.Flags().StringVarP(&profile, "profile", "", "", "Select evaluation profile (basic, extended, additional)")
}
