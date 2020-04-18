package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ping struct {
	seq      int
	interval int
	IP       net.IP //change to IP && add from string
	from     string
	conn     *icmp.PacketConn
}

func (p *ping) sendEcho() {
	gPing.seq++

	time := []byte((time.Now().Format(time.UnixDate)))
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  gPing.seq,
			Data: time,
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := p.conn.WriteTo(wb, &net.IPAddr{IP: p.IP}); err != nil {
		log.Fatal(err)
	}
}

func (p *ping) getResponce() {
	buff := make([]byte, 1500)
	read, _, err := p.conn.ReadFrom(buff)
	if err != nil {
		log.Fatal(err)
	}
	message, err := icmp.ParseMessage(1, buff[:read])
	if err != nil {
		log.Fatal(err)
	}

	if message.Type != ipv4.ICMPTypeEchoReply {
		return
	}

	echoReply := message.Body.(*icmp.Echo)
	RTT := getRoundTripDelay(string(echoReply.Data))

	fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%v\n", read, p.from, echoReply.Seq, 0, RTT)
}

func getRoundTripDelay(lastTime string) time.Duration {

	prevTime, err := time.Parse(time.UnixDate, lastTime)
	if err != nil {
		log.Fatal(err)
	}
	return time.Now().Sub(prevTime)
}

func (p *ping) DNS(addr string) {
	IP, err := net.LookupIP(addr)
	if err != nil {
		log.Fatal(err)
	}
	p.IP = IP[0]

	names, err := net.LookupAddr(IP[0].String())
	if err != nil {
		fmt.Println(err)
	}

	if addr == IP[0].String() {
		p.from = addr
	} else {
		p.from = fmt.Sprintf("%s (%s)", names[0], IP[0].String())
	}

	fmt.Printf("PING %s (%v) --- bytes of data\n", addr, IP[0])
}
