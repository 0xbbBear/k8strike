package core

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"k8strike/pkg/spider/define"
	spiderDNS "k8strike/pkg/spider/dns"
	"k8strike/pkg/spider/metrics"
	"k8strike/pkg/spider/metrics/coredns"
	kubeStateMetrics "k8strike/pkg/spider/metrics/kube-state-metrics"
	"k8strike/pkg/spider/multi"
	"k8strike/pkg/spider/nfs"
	"k8strike/pkg/spider/printer"
	"k8strike/pkg/spider/scanner"
	"k8strike/pkg/tool/kubectl"

	miekgDNS "github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var DnsTimeout int
var Latency int
var LockerMode bool
var NetResolver *spiderDNS.SpiderResolver

var Opts = struct {
	Cidr    string
	PodCidr string

	DnsServer  string
	Zone       string
	OutputFile string
	Verbose    int

	MultiThreadingMode bool
	ThreadingNum       int

	SkipKubeDNSCheck bool

	FilterRules   []string
	FilterStrings []string

	DnsTimeout int
	Latency    int

	DnsutilsDomain string
	DnsutilsType   string

	MetricsURL    string
	MetricsSource string

	NfsHost   string
	NfsPath   string
	NfsAction string
	NfsFile   string
}{}

func defaultPodCidr() string {
	return "10.0.0.1/16"
}

func defaultCidr() string {
	return "10.96.0.1/16"
}

func InitSpiderCommands() {
	InitAllCmd()
	InitAxfrCmd()
	InitWildcardCmd()
	InitNeighborCmd()
	InitWhereisdnsCmd()
	InitDnssdCmd()
	InitDnsutilsCmd()
	InitMetricsCmd()
	InitNfsCmd()
	InitAdmissionCmd()
}

var AllCmd *cobra.Command
var AxfrCmd *cobra.Command
var WildcardCmd *cobra.Command
var NeighborCmd *cobra.Command
var WhereisdnsCmd *cobra.Command
var DnssdCmd *cobra.Command
var DnsutilsCmd *cobra.Command
var MetricsCmd *cobra.Command
var NfsCmd *cobra.Command
var AdmissionCmd *cobra.Command

func InitAllCmd() {
	AllCmd = &cobra.Command{
		Use:   "all",
		Short: "Run all k8s spider modules",
		Run: func(cmd *cobra.Command, args []string) {
			if Opts.Cidr == "" {
				log.Warn("cidr is required")
				return
			}
			zone := Opts.Zone
			if zone == "" {
				zone = "cluster.local"
			}

			records := scanner.DumpWildCard(zone)
			if records != nil {
				printer.PrintResult(records, Opts.OutputFile)
			}

			records, err := scanner.DumpAXFR(miekgDNS.Fqdn(zone), "ns.dns."+zone+":53")
			if err == nil {
				printer.PrintResult(records, Opts.OutputFile)
			}

			ipNets, err := spiderDNS.ParseStringToIPNet(Opts.Cidr)
			if err != nil {
				log.Warnf("ParseStringToIPNet failed: %v", err)
				return
			}
			podNets, err := spiderDNS.ParseStringToIPNet(Opts.PodCidr)
			if err != nil {
				log.Warnf("ParseStringToIPNet failed: %v", err)
				return
			}

			finalRecord := RunMultiThread(ipNets, podNets, Opts.ThreadingNum)
			printer.PrintResult(finalRecord, Opts.OutputFile)
		},
	}
	AddSpiderFlags(AllCmd)
	AllCmd.Flags().BoolVarP(&Opts.MultiThreadingMode, "only-service", "O", false, "only dump service cidr")
}

func InitAxfrCmd() {
	AxfrCmd = &cobra.Command{
		Use:   "axfr",
		Short: "DNS zone transfer",
		Run: func(cmd *cobra.Command, args []string) {
			zone := Opts.Zone
			if zone == "" {
				log.Warn("zone can't empty")
				return
			}
			fqdn := miekgDNS.Fqdn(zone)
			dnsServer := Opts.DnsServer
			if dnsServer == "" {
				dnsServer = "ns.dns." + zone + ":53"
			}
			records, err := scanner.DumpAXFR(fqdn, dnsServer)
			if err != nil {
				log.Errorf("Transfer failed: %v", err)
				return
			}
			printer.PrintResult(records, Opts.OutputFile)
		},
	}
	AddSpiderFlags(AxfrCmd)
}

func InitWildcardCmd() {
	WildcardCmd = &cobra.Command{
		Use:   "wildcard",
		Short: "DNS wildcard detection",
		Run: func(cmd *cobra.Command, args []string) {
			zone := Opts.Zone
			if zone == "" {
				zone = "cluster.local"
			}
			records := scanner.DumpWildCard(zone)
			printer.PrintResult(records, Opts.OutputFile)
		},
	}
	AddSpiderFlags(WildcardCmd)
}

func InitNeighborCmd() {
	NeighborCmd = &cobra.Command{
		Use:   "neighbor",
		Short: "Pod neighbor discovery",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	AddSpiderFlags(NeighborCmd)

	NeighborSvcCmd := &cobra.Command{
		Use:   "svc",
		Short: "Discover neighbor services",
		Run: func(cmd *cobra.Command, args []string) {
			ipNets, err := spiderDNS.ParseStringToIPNet(Opts.PodCidr)
			if err != nil {
				log.Warnf("ParseStringToIPNet failed: %v", err)
				return
			}
			scan := multi.ScanNeighborSvc(ipNets, Opts.ThreadingNum)
			var finalRecord define.Records
			for r := range scan {
				finalRecord = append(finalRecord, r)
			}
			printer.PrintResult(finalRecord, Opts.OutputFile)
		},
	}
	AddSpiderFlags(NeighborSvcCmd)

	NeighborPodCmd := &cobra.Command{
		Use:   "pod",
		Short: "Discover neighbor pods",
		Run: func(cmd *cobra.Command, args []string) {
			ipNets, err := spiderDNS.ParseStringToIPNet(Opts.PodCidr)
			if err != nil {
				log.Warnf("ParseStringToIPNet failed: %v", err)
				return
			}
			scan := multi.ScanNeighbor(nil, ipNets, Opts.ThreadingNum)
			var finalRecord define.Records
			for r := range scan {
				finalRecord = append(finalRecord, r...)
			}
			printer.PrintResult(finalRecord, Opts.OutputFile)
		},
	}
	AddSpiderFlags(NeighborPodCmd)

	NeighborCmd.AddCommand(NeighborSvcCmd, NeighborPodCmd)
}

func InitWhereisdnsCmd() {
	WhereisdnsCmd = &cobra.Command{
		Use:   "whereisdns",
		Short: "Find DNS server",
		Run: func(cmd *cobra.Command, args []string) {
			if spiderDNS.CheckKubeDNS() {
				log.Info("Kubernetes DNS detected")
			} else {
				log.Warn("Kubernetes DNS not detected")
			}
		},
	}
	AddSpiderFlags(WhereisdnsCmd)
}

func InitDnssdCmd() {
	DnssdCmd = &cobra.Command{
		Use:   "dnssd",
		Short: "DNS Service Discovery",
		Run: func(cmd *cobra.Command, args []string) {
			ipNets, err := spiderDNS.ParseStringToIPNet(Opts.Cidr)
			if err != nil {
				log.Warnf("ParseStringToIPNet failed: %v", err)
				return
			}
			scan := multi.ScanAll(ipNets, Opts.ThreadingNum)
			var finalRecord define.Records
			for r := range scan {
				finalRecord = append(finalRecord, r)
			}
			printer.PrintResult(finalRecord, Opts.OutputFile)
		},
	}
	AddSpiderFlags(DnssdCmd)
}

func InitDnsutilsCmd() {
	DnsutilsCmd = &cobra.Command{
		Use:   "dnsutils [domain]",
		Short: "DNS utility operations",
		Run: func(cmd *cobra.Command, args []string) {
			domain := Opts.DnsutilsDomain
			if domain == "" && len(args) > 0 {
				domain = args[0]
			}
			if domain == "" {
				log.Warn("domain is required")
				return
			}

			resolver := spiderDNS.DefaultResolver()
			if Opts.DnsServer != "" {
				resolver = spiderDNS.WarpDnsServer(Opts.DnsServer)
			}

			qtype := strings.ToLower(Opts.DnsutilsType)
			types := []string{qtype}
			if qtype == "all" {
				types = []string{"a", "ptr", "srv", "txt"}
			}

			for _, t := range types {
				switch t {
				case "a":
					ips, err := resolver.ARecord(domain)
					if err != nil {
						log.Warnf("A record query failed: %v", err)
					} else {
						for _, ip := range ips {
							fmt.Printf("A\t%s\t%s\n", domain, ip.String())
						}
					}
				case "ptr":
					ip := net.ParseIP(domain)
					if ip == nil {
						log.Warnf("PTR query requires an IP address, got: %s", domain)
						continue
					}
					names := resolver.PTRRecord(ip)
					for _, name := range names {
						fmt.Printf("PTR\t%s\t%s\n", domain, name)
					}
				case "srv":
					cname, srvs, err := resolver.SRVRecord(domain)
					if err != nil {
						log.Warnf("SRV record query failed: %v", err)
					} else {
						for _, srv := range srvs {
							fmt.Printf("SRV\t%s\t%s\t%d\t%s\n", domain, srv.Target, srv.Port, cname)
						}
					}
				case "txt":
					txts, err := resolver.TXTRecord(domain)
					if err != nil {
						log.Warnf("TXT record query failed: %v", err)
					} else {
						for _, txt := range txts {
							fmt.Printf("TXT\t%s\t%s\n", domain, txt)
						}
					}
				default:
					log.Warnf("unknown query type: %s", t)
				}
			}
		},
	}
	AddSpiderFlags(DnsutilsCmd)
	DnsutilsCmd.Flags().StringVarP(&Opts.DnsutilsDomain, "domain", "D", "", "domain to query")
	DnsutilsCmd.Flags().StringVarP(&Opts.DnsutilsType, "type", "T", "all", "query type: a, ptr, srv, txt, all")
}

func InitMetricsCmd() {
	MetricsCmd = &cobra.Command{
		Use:   "metrics",
		Short: "Metrics endpoint parser",
		Run: func(cmd *cobra.Command, args []string) {
			url := Opts.MetricsURL
			if url == "" {
				endpoints := []string{
					"http://kube-state-metrics.kube-system.svc.cluster.local:8080/metrics",
					"http://kube-state-metrics.monitoring.svc.cluster.local:8080/metrics",
					"http://coredns.kube-system.svc.cluster.local:9153/metrics",
					"http://kube-dns.kube-system.svc.cluster.local:9153/metrics",
				}
				for _, ep := range endpoints {
					if body, err := fetchMetrics(ep); err == nil && body != "" {
						url = ep
						log.Infof("metrics endpoint found: %s", ep)
						break
					}
				}
			}
			if url == "" {
				log.Warn("metrics URL is required, use --url or ensure a metrics endpoint is reachable")
				return
			}

			body, err := fetchMetrics(url)
			if err != nil {
				log.Errorf("failed to fetch metrics: %v", err)
				return
			}

			source := strings.ToLower(Opts.MetricsSource)
			if source == "" || source == "auto" {
				if strings.Contains(body, "coredns_plugin_enabled") {
					source = "coredns"
				} else if strings.Contains(body, "kube_pod_container_info") {
					source = "kube-state-metrics"
				}
			}

			var rules metrics.MatchRules
			switch source {
			case "coredns":
				rules = coredns.CoreDNSMatchRules()
			case "kube-state-metrics":
				rules = kubeStateMetrics.DefaultMatchRules()
			default:
				rules = append(kubeStateMetrics.DefaultMatchRules(), coredns.CoreDNSMatchRules()...)
			}

			if err := rules.Compile(); err != nil {
				log.Errorf("compile matcher failed: %v", err)
				return
			}

			var matched []*metrics.MetricMatcher
			scanner := bufio.NewScanner(strings.NewReader(body))
			for scanner.Scan() {
				line := scanner.Text()
				res, err := rules.Match(line)
				if err != nil {
					continue
				}
				matched = append(matched, res.CopyData())
			}

			resources := metrics.ConvertToResource(matched)
			if len(resources) == 0 {
				log.Warn("no metrics matched")
				return
			}

			var rl define.ResourceList = resources
			if Opts.OutputFile != "" {
				f, err := os.OpenFile(Opts.OutputFile, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Warnf("open output file failed: %v", err)
					return
				}
				defer f.Close()
				rl.Print(f)
			} else {
				rl.Print()
			}
		},
	}
	AddSpiderFlags(MetricsCmd)
	MetricsCmd.Flags().StringVarP(&Opts.MetricsURL, "url", "u", "", "metrics endpoint URL")
	MetricsCmd.Flags().StringVarP(&Opts.MetricsSource, "source", "s", "auto", "metrics source: coredns, kube-state-metrics, auto")
}

func InitNfsCmd() {
	NfsCmd = &cobra.Command{
		Use:   "nfs",
		Short: "NFS operations",
		Run: func(cmd *cobra.Command, args []string) {
			host := Opts.NfsHost
			path := Opts.NfsPath
			if host == "" || path == "" {
				log.Warn("nfs --host and --path are required")
				return
			}

			client, err := nfs.NewNFSClient(host, path, nfs.NewRootAuth())
			if err != nil {
				log.Errorf("connect to NFS failed: %v", err)
				return
			}
			defer client.Close()

			action := strings.ToLower(Opts.NfsAction)
			switch action {
			case "ls", "list":
				entries, err := client.ListDir(Opts.NfsFile)
				if err != nil {
					log.Errorf("list dir failed: %v", err)
					return
				}
				for _, entry := range entries {
					fmt.Printf("%s\n", entry.Name())
				}
			case "cat":
				data, err := client.Cat(Opts.NfsFile)
				if err != nil {
					log.Errorf("cat file failed: %v", err)
					return
				}
				fmt.Print(string(data))
			case "stat":
				info, err := client.Stat(Opts.NfsFile)
				if err != nil {
					log.Errorf("stat file failed: %v", err)
					return
				}
				fmt.Printf("Name: %s\nSize: %d\nMode: %v\nModTime: %v\n",
					info.Name(), info.Size(), info.Mode(), info.ModTime())
			default:
				log.Warnf("unknown action: %s, supported: ls, cat, stat", action)
			}
		},
	}
	AddSpiderFlags(NfsCmd)
	NfsCmd.Flags().StringVarP(&Opts.NfsHost, "host", "H", "", "NFS server host")
	NfsCmd.Flags().StringVarP(&Opts.NfsPath, "path", "P", "", "NFS export path")
	NfsCmd.Flags().StringVarP(&Opts.NfsAction, "action", "a", "ls", "action: ls, cat, stat")
	NfsCmd.Flags().StringVar(&Opts.NfsFile, "file", ".", "target file or directory")
}

func InitAdmissionCmd() {
	AdmissionCmd = &cobra.Command{
		Use:   "admission",
		Short: "Admission webhook exploit",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := kubectl.ServerAccountRequest(kubectl.K8sRequestOption{
				Api:    "/apis/admissionregistration.k8s.io/v1/validatingwebhookconfigurations",
				Method: "GET",
			})
			if err != nil {
				log.Warnf("failed to get validatingwebhookconfigurations: %v", err)
			} else {
				printWebhooks("ValidatingWebhook", resp)
			}

			resp, err = kubectl.ServerAccountRequest(kubectl.K8sRequestOption{
				Api:    "/apis/admissionregistration.k8s.io/v1/mutatingwebhookconfigurations",
				Method: "GET",
			})
			if err != nil {
				log.Warnf("failed to get mutatingwebhookconfigurations: %v", err)
			} else {
				printWebhooks("MutatingWebhook", resp)
			}
		},
	}
	AddSpiderFlags(AdmissionCmd)
}

func fetchMetrics(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func printWebhooks(kind, jsonStr string) {
	items := gjson.Get(jsonStr, "items")
	if !items.Exists() || len(items.Array()) == 0 {
		log.Infof("no %s configurations found", kind)
		return
	}
	for _, item := range items.Array() {
		name := item.Get("metadata.name").String()
		fmt.Printf("%s: %s\n", kind, name)
		webhooks := item.Get("webhooks")
		for _, wh := range webhooks.Array() {
			whName := wh.Get("name").String()
			svcName := wh.Get("clientConfig.service.name").String()
			svcNs := wh.Get("clientConfig.service.namespace").String()
			svcPath := wh.Get("clientConfig.service.path").String()
			if svcName != "" {
				fmt.Printf("  webhook: %s -> service: %s/%s%s\n", whName, svcNs, svcName, svcPath)
			} else {
				url := wh.Get("clientConfig.url").String()
				fmt.Printf("  webhook: %s -> url: %s\n", whName, url)
			}
		}
	}
}

func AddSpiderFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&Opts.Cidr, "cidr", "c", defaultCidr(), "cidr like: 192.168.0.1/16")
	cmd.Flags().StringVarP(&Opts.PodCidr, "pod-cidr", "p", defaultPodCidr(), "pod cidr")
	cmd.Flags().StringVarP(&Opts.DnsServer, "dns-server", "d", "", "dns server")
	cmd.Flags().IntVarP(&Opts.DnsTimeout, "dns-timeout", "i", 2, "dns timeout")
	cmd.Flags().StringVarP(&Opts.Zone, "zone", "z", "cluster.local", "zone")
	cmd.Flags().StringVarP(&Opts.OutputFile, "output-file", "o", "", "output file")
	cmd.Flags().CountVarP(&Opts.Verbose, "verbose", "v", "log level (-v debug,-vv trace)")
	cmd.Flags().BoolVarP(&Opts.MultiThreadingMode, "thread", "t", false, "multi threading mode")
	cmd.Flags().IntVarP(&Opts.ThreadingNum, "thread-num", "n", 16, "threading num")
	cmd.Flags().BoolVarP(&Opts.SkipKubeDNSCheck, "skip-kube-dns-check", "k", false, "skip kube-dns check")
	cmd.Flags().StringSliceVarP(&Opts.FilterRules, "filter-rules", "F", []string{}, "filter regexp rules")
	cmd.Flags().StringSliceVarP(&Opts.FilterStrings, "filter-strings", "f", []string{}, "filter contained strings")
	cmd.Flags().IntVarP(&Opts.Latency, "latency", "l", 0, "Latency control while each dns query in ms")
	cmd.PersistentFlags().SortFlags = false
}

func RunMultiThread(net, pod *net.IPNet, count int) (finalRecord define.Records) {
	scan := multi.ScanAll(net, count)
	scan2 := multi.ScanNeighborSvc(pod, count)
	for r := range mergeRecords(scan, scan2) {
		finalRecord = append(finalRecord, r)
	}
	return
}

func mergeRecords(cs ...<-chan define.Record) chan define.Record {
	out := make(chan define.Record)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan define.Record) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
