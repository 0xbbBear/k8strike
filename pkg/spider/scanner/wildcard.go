package scanner

import (
	"k8strike/pkg/spider/define"
	spiderDNS "k8strike/pkg/spider/dns"
	miekgDNS "github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

func DumpWildCard(zone string) []define.Record {
	searchDNS := []string{
		miekgDNS.Fqdn("any.any.svc." + zone),
		miekgDNS.Fqdn("any.any.any.svc." + zone),
	}
	var records []define.Record
	for _, dnsDomain := range searchDNS {
		_, srv, err := spiderDNS.SRVRecord(dnsDomain)
		if err != nil {
			log.Warnf("wildcard dns query to %v failed: %v", dnsDomain, err)
			continue
		}
		r := define.Record{}
		r.SetSrvRecord(dnsDomain, srv)
		records = append(records, r)
	}
	return records
}
