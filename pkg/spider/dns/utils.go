package dns

import (
	log "github.com/sirupsen/logrus"
)

func CheckPodVerified() bool {
	iplist := []string{
		"8.8.8.8",
		"1.1.1.1",
		"114.114.114.114",
	}
	for _, ip := range iplist {
		targetHostName := IPtoPodHostName(ip, "kube-system")
		log.Tracef("test if record %v is ip: %v", targetHostName, ip)
		ips, err := ARecord(targetHostName)
		if err != nil {
			continue
		}
		for _, i := range ips {
			if i.String() == ip {
				return false
			}
		}
	}
	return true
}

func CheckKubeDNS(dns ...*SpiderResolver) bool {
	if len(dns) > 1 {
		for _, d := range dns {
			if CheckKubeDNS(d) {
				log.Infof("kubernetes cluster found in dns(%v)", d.CurrentDNS())
				return true
			}
		}
	} else {
		rs := NetResolver
		if rs == nil {
			rs = DefaultResolver()
		}
		if len(dns) > 0 {
			rs = dns[0]
		}
		if CheckKubeDNS_DefaultAPIServer(rs) || CheckKubeDNS_DNSVersion(rs) || CheckKubeDNS_NS_DNS_DOMAIN(rs) {
			return true
		}
	}
	return false
}

func CheckKubeDNS_DefaultAPIServer(dns *SpiderResolver) bool {
	info, err := dns.ARecord("kubernetes.default.svc." + Zone)
	if err == nil {
		log.Debugf("kubernetes.default.svc.%v found in dns(%v)! response: %v", Zone, dns.CurrentDNS(), info)
		return true
	}
	log.Tracef("kubernetes.default.svc.%v not found in dns(%v)", Zone, dns.CurrentDNS())
	info, err = dns.ARecord("kubernetes.default.svc")
	if err == nil {
		log.Warnf("kubernetes.default.svc found in dns(%v)! response: %v, maybe %v is incorrect", dns.CurrentDNS(), info, Zone)
		return true
	}
	log.Tracef("kubernetes.default.svc not found in dns(%v)", dns.CurrentDNS())
	return false
}

func CheckKubeDNS_DNSVersion(dns *SpiderResolver) bool {
	info, err := dns.TXTRecord("dns-version." + Zone)
	if err == nil {
		log.Debugf("dns-version.%v found in dns(%v)! response: %v", Zone, dns.CurrentDNS(), info)
		return true
	}
	log.Tracef("dns-version.%v not found in dns(%v)", Zone, dns.CurrentDNS())
	info, err = dns.TXTRecord("dns-version")
	if err == nil {
		log.Warnf("dns-version found in dns(%v)! response: %v, maybe %v is incorrect", dns.CurrentDNS(), info, Zone)
		return true
	}
	log.Tracef("dns-version not found in dns(%v)", dns.CurrentDNS())
	return false
}

func CheckKubeDNS_NS_DNS_DOMAIN(dns *SpiderResolver) bool {
	info, err := dns.ARecord("ns.dns." + Zone)
	if err == nil {
		log.Debugf("ns.dns.%v found in dns(%v)! response: %v", Zone, dns.CurrentDNS(), info)
		return true
	}
	log.Tracef("ns.dns.%v not found in dns(%v)", Zone, dns.CurrentDNS())
	info, err = dns.ARecord("ns.dns")
	if err == nil {
		log.Warnf("ns.dns found in dns(%v)! response: %v, maybe %v is incorrect", dns.CurrentDNS(), info, Zone)
		return true
	}
	log.Tracef("ns.dns not found in dns(%v)", dns.CurrentDNS())
	return false
}
