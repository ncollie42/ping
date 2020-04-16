package main

import (
	"log"
	"net"

	"golang.org/x/net/ipv4"
)

func main() {
	protocol := "icmp"
	netaddr, _ := net.ResolveIPAddr("ip4", "0.0.0.0")
	conn, err := net.ListenIP("ip4:"+protocol, netaddr)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		buf := make([]byte, 1024)

		numRead, _, _ := conn.ReadFrom(buf)
		if buf[0] == 3 {
			log.Println("ICMP code 3 -- Port Unreachable")
		} else {
			log.Println("\n----------ICMP--------- len:", numRead)
			log.Printf("\ntype: %v\ncode: %d\nchecksum: % X\nID: % X\nSeq: % X\nData: %s\n", ipv4.ICMPType(buf[0]), buf[1], buf[2:4], buf[4:6], buf[6:8], buf[8:])
			log.Printf("Hex: % X\n", buf[:numRead])
		}
	}

}
