package cli

import (
	"k8strike/pkg/tool/dockerd_api"
	"k8strike/pkg/tool/etcdctl"
	"k8strike/pkg/tool/kubectl"
	"k8strike/pkg/tool/netcat"
	"k8strike/pkg/tool/netstat"
	"k8strike/pkg/tool/network"
	"k8strike/pkg/tool/probe"
	"k8strike/pkg/tool/ps"
	"strconv"

	"github.com/spf13/cobra"
)

var ToolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Built-in tools for container penetration",
	Long:  `Various built-in tools for container penetration testing: ps, netstat, ifconfig, probe, nc, kcurl, ectl, ucurl, dcurl`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "Show process information like ps -ef",
		Run: func(cmd *cobra.Command, args []string) {
			ps.RunPs()
		},
	}
	ToolCmd.AddCommand(psCmd)

	netstatCmd := &cobra.Command{
		Use:   "netstat",
		Short: "Show network connections like netstat -antup",
		Run: func(cmd *cobra.Command, args []string) {
			netstat.RunNetstat()
		},
	}
	ToolCmd.AddCommand(netstatCmd)

	ifconfigCmd := &cobra.Command{
		Use:   "ifconfig",
		Short: "Show network interface information",
		Run: func(cmd *cobra.Command, args []string) {
			network.GetLocalAddresses()
		},
	}
	ToolCmd.AddCommand(ifconfigCmd)

	probeCmd := &cobra.Command{
		Use:   "probe <ip> <port> <parallel> <timeout-ms>",
		Short: "TCP port scan, example: k8strike tool probe 10.0.1.0-255 80,8080-9443 50 1000",
		Args:  cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			ip := args[0]
			portRange := args[1]
			parallel, _ := strconv.ParseInt(args[2], 10, 64)
			timeoutMS, _ := strconv.Atoi(args[3])
			probe.TCPScanToolAPI(ip, portRange, parallel, timeoutMS)
		},
	}
	ToolCmd.AddCommand(probeCmd)

	ncCmd := &cobra.Command{
		Use:   "nc",
		Short: "Netcat implementation for TCP/UDP tunnels",
		Run: func(cmd *cobra.Command, args []string) {
			for _, arg := range args {
				if arg == "-h" || arg == "-help" || arg == "--help" {
					netcat.PrintUsage()
					return
				}
			}
			netcat.RunNetcatWithArgs(args)
		},
	}
	ncCmd.DisableFlagParsing = true
	ToolCmd.AddCommand(ncCmd)

	kcurlCmd := &cobra.Command{
		Use:   "kcurl <token_path> <get|post> <url> [<data>]",
		Short: "Make request to K8s api-server",
		Long: `kcurl - send HTTP request to K8s api-server.

Usage:
  k8strike tool kcurl (<token_path>|anonymous|default) (get|post) <url> [<post_data>]

Options:
  token_path  connect api-server with user-specified service-account token.
  anonymous   connect api-server using system:anonymous service-account.
  default     connect api-server using pod default service-account token.`,
		Args: cobra.RangeArgs(3, 4),
		Run: func(cmd *cobra.Command, args []string) {
			kubectl.KubectlToolApi(args)
		},
	}
	ToolCmd.AddCommand(kcurlCmd)

	ectlCmd := &cobra.Command{
		Use:   "ectl <endpoint> <get> <key>",
		Short: "Unauthorized enumeration of etcd keys",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			etcdctl.EtcdctlToolApi(args)
		},
	}
	ToolCmd.AddCommand(ectlCmd)

	ucurlCmd := &cobra.Command{
		Use:   "ucurl <get|post> <socket> <uri> [<data>]",
		Short: "Make request to docker unix socket",
		Args:  cobra.RangeArgs(3, 4),
		Run: func(cmd *cobra.Command, args []string) {
			dockerd_api.UcurlToolApi(args)
		},
	}
	ToolCmd.AddCommand(ucurlCmd)

	dcurlCmd := &cobra.Command{
		Use:   "dcurl <get|post> <url> [<data>]",
		Short: "Docker daemon curl",
		Args:  cobra.RangeArgs(2, 3),
		Run: func(cmd *cobra.Command, args []string) {
			dockerd_api.DcurlToolApi(args)
		},
	}
	ToolCmd.AddCommand(dcurlCmd)
}