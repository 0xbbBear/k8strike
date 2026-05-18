package util

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	file  = "/proc/net/route"
	line  = 1
	sep   = "\t"
	field = 2
)

func GetGateway() (string, error) {

	file, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		for i := 0; i < line; i++ {
			scanner.Scan()
		}

		tokens := strings.Split(scanner.Text(), sep)
		gatewayHex := "0x" + tokens[field]

		d, _ := strconv.ParseInt(gatewayHex, 0, 64)
		d32 := uint32(d)

		ipd32 := make(net.IP, 4)
		binary.LittleEndian.PutUint32(ipd32, d32)

		ip := net.IP(ipd32).String()

		return ip, nil
	}

	return "", fmt.Errorf("no default gateway found")
}
