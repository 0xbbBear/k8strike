package dns

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

func ParseStringToIPNet(s string) (ipnet *net.IPNet, err error) {
	_, ipnet, err = net.ParseCIDR(s)
	return
}

func ParseIPNetToIPs(ipv4Net *net.IPNet) (ips []net.IP) {
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	finish := (start & mask) | (mask ^ 0xffffffff)

	for i := start; i <= finish; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		ips = append(ips, ip)
	}
	return
}

func IPtoPodHostName(ip, namespace string) string {
	return fmt.Sprintf("%s.%s.pod.%s", strings.ReplaceAll(ip, ".", "-"), namespace, Zone)
}
