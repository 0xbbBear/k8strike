package printer

import (
	"os"

	"k8strike/pkg/spider/define"
	log "github.com/sirupsen/logrus"
)

func PrintResult(records define.Records, OutputFile string) {
	if OutputFile != "" {
		f, err := os.OpenFile(OutputFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Warnf("OpenFile failed: %v", err)
		}
		defer f.Close()
		records.Print(f)
	} else {
		records.Print()
	}
}
