package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const IPheaderSize = 20
const ICMPheaderSize = 8

func init() {
	runtime := runtime.GOOS

	if runtime != "linux" && runtime != "darwin" {
		fmt.Println(runtime, " is not suported")
		os.Exit(1)
	}
	if os.Getegid() != 0 {
		fmt.Println("ping must be ran with sudo")
		os.Exit(1)
	}
}

func main() {
	destination, intervals, size, ttl := flags()

	localIP := localIP()
	if localIP == "" {
		fmt.Println("Could not find IPv4")
		os.Exit(1)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", localIP)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.IPv4PacketConn().SetTTL(*ttl)
	if err != nil {
		log.Fatal(err)
	}

	targetIP, fromStr := doDNS(destination)

	fmt.Printf("PING %s (%s) %d(%d) bytes of data\n", destination, targetIP, *size, *size+ICMPheaderSize+IPheaderSize)
	
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Duration(*intervals) * time.Second).C
	stats := stats{start: time.Now()}
	go func() {
		seq := 1
		for {
			select {
			case <-interrupt:
				statisticsAndQuit(stats, destination)
			case <-ticker:
				sendEcho(conn, seq, targetIP, &stats.sentCount, *size)

			}
			seq++
		}
	}()
	for {
		getResponce(conn, fromStr, &stats, *ttl)
	}

	fmt.Println("Hello world", intervals, localIP)
}


func sendEcho(conn *icmp.PacketConn, seq int, IP net.IP, sentCount *int, payloadSize int) {
	time := []byte((time.Now().Format(time.UnixDate)))
	padding := []byte{}
	if payloadSize > len(time) {
		padding = make([]byte, payloadSize-len(time))
	}
	data := append(time, padding...)
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: data,
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := conn.WriteTo(wb, &net.IPAddr{IP: IP}); err != nil {
		log.Fatal(err)
	}
	(*sentCount)++
}

func getResponce(conn *icmp.PacketConn, fromStr string, stats *stats, ttl int) {
	buff := make([]byte, 1500)
	read, _, err := conn.ReadFrom(buff)
	if err != nil {
		log.Fatal(err)
	}
	message, err := icmp.ParseMessage(1, buff[:read])
	if err != nil {
		log.Fatal(err)
	}

	tailMsg := ""
	seq := 0
	switch reply := message.Body.(type) {
	case *icmp.Echo:
		RTT := roundTripDelay(string(reply.Data))
		stats.RTT = append(stats.RTT, RTT)
		tailMsg = fmt.Sprintf("ttl=%d time=%v", ttl, RTT)
		seq = reply.Seq
		stats.resvCount++
	case *icmp.TimeExceeded:
		tailMsg = fmt.Sprint("Time to live exceeded")
		seq = int(reply.Data[27])
		stats.errsCount++
	default:
		return
	}

	fmt.Printf("%d bytes from %s: icmp_seq=%d %s\n", read, fromStr, seq, tailMsg)
}

func roundTripDelay(lastTime string) time.Duration {
	lastTime = strings.TrimRight(lastTime, string(byte(0)))
	prevTime, err := time.Parse(time.UnixDate, lastTime)
	if err != nil {
		log.Fatal(err)
	}
	return time.Now().Sub(prevTime)
}
