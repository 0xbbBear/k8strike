package scanner

import (
	"net"

	"k8strike/pkg/spider/define"
	spiderDNS "k8strike/pkg/spider/dns"
	log "github.com/sirupsen/logrus"
)

func ScanSingleIP(subnet net.IP) (records []define.Record) {
	ptr := spiderDNS.PTRRecord(subnet)
	if len(ptr) > 0 {
		for _, domain := range ptr {
			log.Infof("PTRrecord %v --> %v", subnet, domain)
			r := define.Record{Ip: subnet, SvcDomain: domain}
			records = append(records, r)
		}
	}
	return
}

func ScanSubnet(subnet *net.IPNet) (records []define.Record) {
	for _, ip := range spiderDNS.ParseIPNetToIPs(subnet) {
		ptr := spiderDNS.PTRRecord(ip)
		if len(ptr) > 0 {
			for _, domain := range ptr {
				log.Infof("PTRrecord %v --> %v", ip, domain)
				r := define.Record{Ip: ip, SvcDomain: domain}
				records = append(records, r)
			}
		} else {
			continue
		}
	}
	return
}

func ScanSingleSvcForPorts(records define.Record) define.Record {
	cname, srv, err := spiderDNS.SRVRecord(records.SvcDomain)
	if err != nil {
		log.Tracef("SRVRecord for %v,failed: %v", records.SvcDomain, err)
		return records
	}
	for _, s := range srv {
		log.Infof("SRVRecord: %v --> %v:%v", records.SvcDomain, s.Target, s.Port)
	}
	records.SetSrvRecord(cname, srv)
	return records
}

func ScanSvcForPorts(records []define.Record) []define.Record {
	for i, r := range records {
		cname, srv, err := spiderDNS.SRVRecord(r.SvcDomain)
		if err != nil {
			log.Tracef("SRVRecord for %v,failed: %v", r.SvcDomain, err)
			continue
		}
		for _, s := range srv {
			log.Infof("SRVRecord: %v --> %v:%v", r.SvcDomain, s.Target, s.Port)
		}
		records[i].SetSrvRecord(cname, srv)
	}
	return records
}
