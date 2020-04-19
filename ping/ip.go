package main

import (
	"fmt"
	"net"
	"os"
)

func doDNS(addr string) (targetIP net.IP, fromStr string) {
	IPs, err := net.LookupIP(addr)
	if err != nil {
		fmt.Println("Lookup: ping: google.com: Name or service not known")
		os.Exit(1)
	}
	for _, IP := range IPs {
		if IP.To4() != nil {
			targetIP = IP
			break
		}
	}
	if targetIP == nil {
		fmt.Println("Destination does not have an IPv4 addr")
		os.Exit(1)
	}

	names, err := net.LookupAddr(targetIP.String())
	if err != nil {
		fmt.Println(err)
	}

	if addr == targetIP.String() {
		fromStr = addr
	} else {
		fromStr = fmt.Sprintf("%s (%s)", names[0], targetIP.String())
	}

	return
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		switch val := addr.(type) {
		case *net.IPNet:
			ip := val.IP
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() == nil {
				continue
			}
			return ip.String()
		}
	}
	return ""
}
