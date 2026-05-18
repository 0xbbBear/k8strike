package multi

import (
	"k8strike/pkg/spider/define"
	"k8strike/pkg/spider/scanner"

	log "github.com/sirupsen/logrus"
)

func ScanServiceWithChan(rev <-chan define.Record) <-chan define.Record {
	out := make(chan define.Record, 100)
	go func() {
		log.Tracef("piped scanning service port begin")
		for records := range rev {
			out <- scanner.ScanSingleSvcForPorts(records)
		}
		close(out)
		log.Tracef("piped scanning service port ends")
	}()
	return out
}
