package post

import (
	"net"
	"strings"

	"k8strike/pkg/spider/define"
	miekgDNS "github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

func RecordsDumpFullService(r []define.Record, zone string) []string {
	var result []string
	for _, record := range r {
		if record.SvcDomain != "" && IsPodServiceFormat(record.SvcDomain, zone) {
			result = append(result, GetPodServiceRawService(record.SvcDomain, zone))
		} else if record.SvcDomain != "" && IsServiceFormat(record.SvcDomain, zone) {
			result = append(result, record.SvcDomain)
		}
		for _, srv := range record.SrvRecords {
			for _, s := range srv.Srv {
				if IsPodServiceFormat(s.Target, zone) {
					result = append(result, GetPodServiceRawService(s.Target, zone))
				} else if IsServiceFormat(s.Target, zone) {
					result = append(result, miekgDNS.Fqdn(s.Target))
				} else {
					log.Debugf("Unhandled service type: %v", s.Target)
				}
			}
		}
		R := ExtraParser(record.Extra)
		if R != "" {
			if IsServiceFormat(R, zone) {
				result = append(result, R)
			}
		}
	}
	return UniqueSlice(result)
}

func IsServiceFormat(domain string, zone string) bool {
	zonelen := len(strings.Split(miekgDNS.Fqdn(zone), "."))
	dn := ReverseSlice(strings.Split(miekgDNS.Fqdn(domain), "."))
	if len(dn) > 4 {
		if dn[zonelen] == "svc" {
			return true
		}
	}
	return false
}

func RecordsDumpNameSpace(r []define.Record, zone string) []string {
	result := RecordsDumpFullService(r, zone)
	for i, record := range result {
		result[i] = GetNamespaceFromDomain(record, zone)
	}
	return UniqueSlice(result)
}

func ExtraParser(record string) string {
	return strings.Split(record, " ")[0]
}

func GetNamespaceFromDomain(domain string, zone string) string {
	zonelen := len(strings.Split(miekgDNS.Fqdn(zone), "."))
	dn := ReverseSlice(strings.Split(miekgDNS.Fqdn(domain), "."))
	if len(dn) > 4 {
		if dn[zonelen] == "svc" {
			return dn[1+zonelen]
		}
	}
	return ""
}

func IsPodServiceFormat(domain, zone string) bool {
	if domain == "" {
		return false
	}
	str := strings.Split(strings.ReplaceAll(miekgDNS.Fqdn(domain), miekgDNS.Fqdn(zone), ""), ".")
	if len(str) > 3 {
		if str[3] == "svc" {
			return true
		}
	}
	return false
}

func GetPodServiceRawService(domain, zone string) string {
	str := strings.Split(strings.ReplaceAll(miekgDNS.Fqdn(domain), miekgDNS.Fqdn(zone), ""), ".")
	return miekgDNS.Fqdn(strings.Join(str[1:], ".") + zone)
}

func GetPodServiceRawIP(domain, zone string) net.IP {
	str := strings.Split(strings.ReplaceAll(miekgDNS.Fqdn(domain), miekgDNS.Fqdn(zone), ""), ".")
	return net.ParseIP(strings.ReplaceAll(str[0], "-", "."))
}

func PodServiceMap(BaseService define.Records, zone string) map[string][]string {
	result := make(map[string][]string)
	for _, r := range BaseService {
		log.Tracef("Processing Record: %v", r)
		if r.SvcDomain != "" && IsPodServiceFormat(r.SvcDomain, zone) {
			svcDomain := GetPodServiceRawService(r.SvcDomain, zone)
			result[svcDomain] = append(result[svcDomain], GetPodServiceRawIP(r.SvcDomain, zone).String())
		} else if r.SvcDomain != "" && IsServiceFormat(r.SvcDomain, zone) {
			if r.Ip != nil {
				result[miekgDNS.Fqdn(r.SvcDomain)] = append(result[miekgDNS.Fqdn(r.SvcDomain)], r.Ip.String())
			} else {
				log.Debugf("Lost service ip addr %v", r.SvcDomain)
			}
		} else {
			log.Debugf("Unhandled service type: %v", r.SvcDomain)
		}

		for _, srv := range r.SrvRecords {
			for _, s := range srv.Srv {
				if IsPodServiceFormat(s.Target, zone) {
					domain := GetPodServiceRawService(s.Target, zone)
					result[domain] = append(result[domain], GetPodServiceRawIP(s.Target, zone).String())
				} else if IsServiceFormat(s.Target, zone) {
					if r.Ip != nil {
						result[miekgDNS.Fqdn(s.Target)] = append(result[miekgDNS.Fqdn(s.Target)], r.Ip.String())
					} else {
						log.Debugf("Lost service ip addr %v because of in srv record", s.Target)
						log.Debugf("can't put ip address in service %v in record %v", miekgDNS.Fqdn(s.Target), r)
					}
				} else {
					log.Debugf("Unhandled service type: %v", s.Target)
				}
			}
		}
	}
	for k, v := range result {
		result[k] = UniqueSlice(v)
	}
	return result
}
