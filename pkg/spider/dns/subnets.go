package dns

import (
	"fmt"
	"math"
	"net"
)

func SubnetInto(network *net.IPNet, count int) ([]*net.IPNet, error) {
	maskBits, _ := network.Mask.Size()
	hostBits := 32 - maskBits
	hostCount := 1 << uint(hostBits)

	ideal := float64(hostCount) / float64(count)
	newHostBits := int(math.Log2(ideal))
	shift := hostBits - newHostBits
	return SubnetShift(network, shift)
}

func SubnetShift(network *net.IPNet, bits int) ([]*net.IPNet, error) {
	if bits < 0 {
		return nil, fmt.Errorf("bit shift may not be negative, got %d", bits)
	}
	if bits > 31 {
		return nil, fmt.Errorf("network subnets cannot be divided %d times", bits)
	}
	subnetCount := 1 << uint(bits)
	subnets := make([]*net.IPNet, subnetCount)

	start := network.IP
	maskBits, _ := network.Mask.Size()
	hostBits := 32 - maskBits

	if maskBits+bits > 32 {
		return nil, fmt.Errorf("network subnet mask greater than /32, /%d is invalid", maskBits+bits)
	}

	newMaskBits := maskBits + bits
	newHostBits := hostBits - bits
	newMask := net.CIDRMask(newMaskBits, 32)

	hostCount := 1 << uint(newHostBits)

	for i := 0; i < subnetCount; i++ {
		ip := numeric(start) + uint32(i*hostCount)
		subnets[i] = &net.IPNet{
			IP:   bytewise(ip),
			Mask: newMask,
		}
	}

	return subnets, nil
}


func numeric(bytes net.IP) uint32 {
	var ip uint32
	for i, b := range []byte(bytes) {
		ip |= uint32(b) << (8 * uint32(3-i))
	}
	return ip
}

func bytewise(numeric uint32) net.IP {
	ip := make([]byte, 4)
	for i := 3; i >= 0; i-- {
		ip[i] = byte(numeric & 0xFF)
		numeric >>= 8
	}
	return net.IP(ip)
}
