package scanner

import (
	"net"

	"k8strike/pkg/spider/dns"

	log "github.com/sirupsen/logrus"
)

func ScanPodExist(ip net.IP, ns string) bool {
	targetHostName := dns.IPtoPodHostName(ip.String(), ns)
	ips, err := dns.ARecord(targetHostName)
	if err != nil {
		log.Tracef("ScanPodExist %v failed: %v", ip.String(), err)
		return false
	}
	for _, i := range ips {
		if i.String() == ip.String() {
			return true
		}
	}
	return false
}
