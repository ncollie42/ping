package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

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

func main() {

	intervals := flag.Int("i", 1, "Wait interval between seding each packet.")
	size := flag.Int("s", 56, " Specifies the number of data bytes to be sent. The default is 56, which translates into 64 ICMP data bytes when combined with the 8 bytes of ICMP header data.")
	ttl := flag.Int("t", 50, "Set the IP Time to Live.")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("ping: usage error: Destination address required")
		os.Exit(1)
	}
	destination := flag.Arg(0)
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

	conn.IPv4PacketConn().SetTTL(*ttl)

	targetIP, fromStr := doDNS(destination)
	fmt.Printf("PING %s (%s) %d(%d) bytes of data\n", destination, targetIP, *size, *size+28)
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

type stats struct {
	sentCount int
	resvCount int
	errsCount int
	RTT       []time.Duration
	start     time.Time
}

func statisticsAndQuit(stats stats, Destination string) {
	packetLoss := (1 - (float64(stats.resvCount) / float64(stats.sentCount))) * 100

	fmt.Println("\n---", Destination, "ping statistics ---")
	errors := ""
	if stats.errsCount > 0 {
		errors = fmt.Sprintf("+%d errors, ", stats.errsCount)
	}
	fmt.Printf("%d packets transmited, %d received, %s%.2f%% packet loss, time %.3fs\n",
		stats.sentCount, stats.resvCount, errors, packetLoss, time.Now().Sub(stats.start).Seconds())

	sort.Sort(byTime(stats.RTT))
	if len(stats.RTT) > 0 {
		min := stats.RTT[0]
		avg := avg(stats.RTT)
		max := stats.RTT[len(stats.RTT)-1]
		mdev := stats.RTT[len(stats.RTT)/2]
		fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n", min.Seconds(), avg.Seconds(), max.Seconds(), mdev.Seconds())
	}
	os.Exit(0)
}

func avg(times []time.Duration) time.Duration {
	tottal := time.Duration(0)
	for _, val := range times {
		tottal += val
	}
	tottal = tottal / time.Duration(len(times))
	return tottal
}

type byTime []time.Duration

func (t byTime) Len() int {
	return len(t)
}

func (t byTime) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t byTime) Less(i, j int) bool {
	return t[i] < t[j]
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
