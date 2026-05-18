package scanner

import (
	"strings"

	"k8strike/pkg/spider/define"
	miekgDNS "github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

func DumpAXFR(target string, dnsServer string) ([]define.Record, error) {
	t := new(miekgDNS.Transfer)
	m := new(miekgDNS.Msg)
	m.SetAxfr(target)
	ch, err := t.In(m, dnsServer)
	if err != nil {
		return nil, err
	}
	var records []define.Record
	for rr := range ch {
		if rr.Error != nil {
			log.Debugf("Error: %v", rr.Error)
			return records, rr.Error
		}
		for _, r := range rr.RR {
			records = append(records, define.Record{
				SvcDomain: r.Header().Name,
				Extra:     strings.Join(strings.Split(r.String(), "\t"), " "),
			})
		}
		log.Debugf("Record: %v", rr.RR)
	}
	return records, nil
}
