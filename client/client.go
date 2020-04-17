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
			checksum := int32(buf[2])<<8 + int32(buf[3])
			ID := int32(buf[4])<<8 + int32(buf[5])
			seq := int32(buf[6])<<8 + int32(buf[6])
			log.Println("----------ICMP--------- len:", numRead)
			log.Println("Type:", ipv4.ICMPType(buf[0]))
			log.Println("Code:", buf[1])
			log.Println("Checksum:", checksum, "Hex:", buf[2:4])
			log.Println("ID:", ID, "Hex:", buf[4:6])
			log.Println("Seq:", seq, "Hex:", buf[6:8])
			log.Println("Data:", string(buf[8:]))
			log.Printf("Full ICMP Hex: % X\n", buf[:numRead])
		}
	}
}
