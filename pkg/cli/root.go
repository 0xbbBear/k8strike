package cli

import (
	"fmt"
	"os"

	"k8strike/pkg/util"

	"github.com/spf13/cobra"
)

var GitCommit string

var Args map[string]interface{}

var BannerTitle = "k8strike"
var BannerVersion = fmt.Sprintf("%s %s", "k8strike Version(GitCommit):", GitCommit)

var BannerHeader = fmt.Sprintf(`%s
%s
Zero-dependency cloudnative k8s/docker/serverless penetration toolkit
`, util.GreenBold.Sprint(BannerTitle), BannerVersion)

var BannerContainerTpl = BannerHeader + `
%s
  k8strike eva
  k8strike eva --full
  k8strike evaluate [--full]
  k8strike run (--list | <exploit> [<args>...])
  k8strike <tool> [<args>...]
  k8strike spider [command]

%s
  k8strike evaluate                              Gather information to find weakness inside container.
  k8strike eva                                  Alias of "k8strike evaluate".
  k8strike evaluate --full                      Enable file scan during information gathering.

%s
  k8strike run --list                           List all available exploits.
  k8strike run <exploit> [<args>...]            Run single exploit, docs in https://github.com/k8strike/k8strike

%s
  k8strike spider all                           Run all k8s spider modules.
  k8strike spider axfr                          AXFR DNS zone transfer.
  k8strike spider wildcard                      DNS wildcard detection.
  k8strike spider neighbor                      Pod neighbor discovery.
  k8strike spider whereisdns                    Find DNS server.
  k8strike spider dnssd                         DNS Service Discovery.
  k8strike spider dnsutils                      DNS utility operations.
  k8strike spider metrics                       Metrics endpoint parser.
  k8strike spider nfs                           NFS operations.
  k8strike spider admission                     Admission webhook exploit.

%s
  ps                                            Show process information like "ps -ef" command.
  netstat                                       Like "netstat -antup" command.
  nc [options]                                  Create TCP tunnel.
  ifconfig                                      Show network information.
  kcurl <path> (get|post) <uri> [<data>]        Make request to K8s api-server.
  ectl <endpoint> get <key>                    Unauthorized enumeration of ectd keys.
  ucurl (get|post) <socket> <uri> <data>       Make request to docker unix socket.
  dcurl <socket> <uri> [<data>]                Docker daemon curl.
  probe <ip> <port> <parallel> <timeout-ms>    TCP port scan, example: k8strike probe 10.0.1.0-255 80,8080-9443 50 1000

%s
  -h --help     Show this help msg.
  -v --version  Show version.
  --profile=<name> Select evaluation profile (basic, extended, additional).
`

var BannerContainer = fmt.Sprintf(
	BannerContainerTpl,
	"Usage:",
	util.GreenBold.Sprint("Evaluate:"),
	util.GreenBold.Sprint("Exploit:"),
	util.GreenBold.Sprint("Spider:"),
	util.GreenBold.Sprint("Tool:"),
	"Options:",
)

var RootCmd = &cobra.Command{
	Use:   "k8strike",
	Short: "k8strike - Cloudnative k8s/docker penetration toolkit",
	Long: `k8strike - Zero-dependency cloudnative k8s/docker/serverless penetration toolkit

Usage: k8strike <command> [flags]
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(BannerContainer)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(EvaluateCmd)
	RootCmd.AddCommand(RunCmd)
	RootCmd.AddCommand(ToolCmd)
	RootCmd.AddCommand(SpiderCmd)
}
