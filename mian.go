package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/thinkboy/log4go"
)

var PAYLOAD_NTP_V2 = []byte{0x17, 0x00, 0x03, 0x2a, 0x00, 0x00, 0x00, 0x00}

func inetAton(ip string) int64 {
	bits := strings.Split(ip, ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func inetNtoa(ip int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ip & 0xFF)
	bytes[1] = byte((ip >> 8) & 0xFF)
	bytes[2] = byte((ip >> 16) & 0xFF)
	bytes[3] = byte((ip >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func readNTPResponse(conn *net.UDPConn) {
	defer conn.Close()
	total := 0
	saddr := ""
	for {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 1024)
		len, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if total > 0 {
				log.Debug("recv UDP from %s len %d", saddr, total)
			}
			return
		}
		saddr = addr.String()
		total += len
	}
}

func main() {
	startTarget := flag.String("s", "192.168.1.1", "start of target")
	endTarget := flag.String("e", "192.168.12.1", "start of target")
	flag.Parse()
	istartTarget := inetAton(*startTarget)
	iendTarget := inetAton(*endTarget)

	for istartTarget <= iendTarget {
		strServerAddr := fmt.Sprintf("%s:123", inetNtoa(istartTarget).String())
		if strings.HasSuffix(strServerAddr, ".1.1:123") {
			log.Debug("Send MONLIST to %s", strServerAddr)
		}
		serverAddr, _ := net.ResolveUDPAddr("udp", strServerAddr)
		conn, _ := net.DialUDP("udp", nil, serverAddr)
		conn.Write(PAYLOAD_NTP_V2)
		go readNTPResponse(conn)
		istartTarget += 1
		time.Sleep(200 * time.Microsecond)
	}
	time.Sleep(5 * time.Second)
}
