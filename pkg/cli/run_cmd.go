package cli

import (
	"fmt"
	"os"

	"k8strike/pkg/plugin"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run exploit modules",
	Long: `Run executes vulnerability exploitation modules against the current environment.
Use --list to see all available exploits.`,
	Run: func(cmd *cobra.Command, args []string) {
		if list {
			plugin.ListAllExploit()
			os.Exit(0)
		}
		if len(args) < 1 {
			cmd.Help()
			return
		}
		name := args[0]
		exploitArgs := args[1:]
		if plugin.Exploits[name] == nil {
			fmt.Printf("\nInvalid script name: %s, available scripts:\n", name)
			plugin.ListAllExploit()
			return
		}
		plugin.RunSingleExploit(name, exploitArgs)
	},
}

var list bool

func init() {
	RunCmd.Flags().BoolVarP(&list, "list", "", false, "List all available exploits")
}
