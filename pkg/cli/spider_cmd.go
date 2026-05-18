package cli

import (
	"k8strike/pkg/spider/core"
	"github.com/spf13/cobra"
)

var SpiderCmd = &cobra.Command{
	Use:   "spider",
	Short: "Kubernetes service discovery tools",
	Long: `Kubernetes service discovery via DNS:
- all: Run all discovery modules
- axfr: DNS zone transfer
- wildcard: DNS wildcard detection
- neighbor: Pod neighbor discovery
- whereisdns: Find DNS server
- dnssd: DNS Service Discovery
- dnsutils: DNS utility operations
- metrics: Metrics endpoint parser
- nfs: NFS operations
- admission: Admission webhook exploit`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	core.InitSpiderCommands()
	SpiderCmd.AddCommand(core.AllCmd)
	SpiderCmd.AddCommand(core.AxfrCmd)
	SpiderCmd.AddCommand(core.WildcardCmd)
	SpiderCmd.AddCommand(core.NeighborCmd)
	SpiderCmd.AddCommand(core.WhereisdnsCmd)
	SpiderCmd.AddCommand(core.DnssdCmd)
	SpiderCmd.AddCommand(core.DnsutilsCmd)
	SpiderCmd.AddCommand(core.MetricsCmd)
	SpiderCmd.AddCommand(core.NfsCmd)
	SpiderCmd.AddCommand(core.AdmissionCmd)
}
