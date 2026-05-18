
package probe

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"k8strike/pkg/conf"
	"golang.org/x/sync/semaphore"
)

type PortScanner struct {
	ipRange   string
	portRange []FromTo
	lock      *semaphore.Weighted
	timeout   time.Duration
}

func ScanPort(ip string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
		}
		return false
	}

	_ = conn.Close()
	return true
}

func (ps *PortScanner) Start() {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	base, start, end, err := GetTaskIPList(ps.ipRange)
	if err != nil {
		log.Println("error found when gene ip list to scan task")
		log.Fatal(err)
	}

	for ipExt := start; ipExt <= end; ipExt++ {
		ip := base + "." + fmt.Sprintf("%d", ipExt)
		for _, p := range ps.portRange {
			for port := p.From; port <= p.To; port++ {
				ps.lock.Acquire(context.TODO(), 1)
				wg.Add(1)
				go func(port int, p FromTo) {
					defer ps.lock.Release(1)
					defer wg.Done()
					if ScanPort(ip, port, ps.timeout) {
						fmt.Printf("open %s: %s:%d\n", p.Desc, ip, port)
					}
				}(port, p)
			}
		}
	}
}

func TCPScanExploitAPI(ipRange string) {
	portFromTo, _ := GetTaskPortList()
	timeout := time.Duration(conf.TCPScannerConf.Timeout) * time.Millisecond

	TCPPScan(ipRange, portFromTo, conf.TCPScannerConf.MaxParallel, timeout)
}

func TCPScanToolAPI(ipRange string, portRange string, parallel int64, timeoutMS int) {
	portFromTo, _ := GetTaskPortListByString(portRange)
	timeout := time.Duration(timeoutMS) * time.Millisecond

	TCPPScan(ipRange, portFromTo, parallel, timeout)
}

func TCPPScan(ipRange string, portRange []FromTo, parallel int64, timeout time.Duration) {

	ps := &PortScanner{
		ipRange:   ipRange,
		portRange: portRange,
		lock:      semaphore.NewWeighted(parallel),
		timeout:   timeout,
	}

	startTime := time.Now()
	log.Printf("scanning %v with user-defined ports, max parallels:%v, timeout:%v\n", ps.ipRange, parallel, ps.timeout)
	ps.Start()

	endTime := time.Now()
	useTime := int64(endTime.Sub(startTime).Seconds() * 1000)
	log.Printf("scanning use time:%vms\n", useTime)
	log.Printf("ending; @args is ips: %v, max parallels:%v, timeout:%v\n", ps.ipRange, conf.TCPScannerConf.MaxParallel, ps.timeout)

}
