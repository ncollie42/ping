package main

import (
	"flag"
	"log"
	"runtime"
	"time"

	"golang.org/x/net/icmp"
)

var gPing *ping

func flags() string {
	destination := flag.String("dest", "N/A", "dns name or ipv4 address")

	flag.IntVar(&gPing.interval, "i", 1, "Wait interval between seding each packet.")

	flag.Parse()
	if *destination == "N/A" {
		log.Fatal("ping: usage error: Destination address required\n\tUse -dest <destination>") //change to not log
	}

	return *destination
}

func init() {
	runtime := runtime.GOOS

	if runtime != "linux" && runtime != "darwin" {
		log.Fatal(runtime, " is not suported")
	}
	gPing = new(ping)

}

func main() {
	var err error
	gPing.conn, err = icmp.ListenPacket("ip4:icmp", "192.168.122.1")
	if err != nil {
		log.Fatal(err)
	}
	defer gPing.conn.Close()

	destination := flags()
	gPing.DNS(destination)

	ticker := time.NewTicker(time.Duration(gPing.interval) * time.Second).C
	go func() {
		for {
			select {
			case <-ticker:
				gPing.sendEcho()
			}
		}
	}()
	for {
		gPing.getResponce()
	}

}

// func main2() {

// 	intervals := flag.Int("i", 1, "Wait interval between seding each packet.")
// 	flag.Parse()

// 	args := flag.Args()
// 	if len(args) != 0 {
// 		fmt.Println("ping: usage error: Destination address required")
// 		os.Exit(1)
// 	}

// 	localIP := localIP()
// 	conn, err := icmp.ListenPacket("ip4:icmp", localIP)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer gPing.conn.Close()

// 	IP, fromStr := newDNS(args[0])

// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)
// 	ticker := time.NewTicker(time.Duration(*intervals) * time.Second).C
// 	sentCount := 0
// 	resvCount := 0
// 	go func() {
// 		seq := 0
// 		for {
// 			select {
// 			case <-interrupt:
// 				statisticsAndQuit(sentCount, resvCount)
// 			case <-ticker:
// 				sendEcho(*conn, seq, IP, &sentCount)

// 			}
// 			seq++
// 		}
// 	}()
// 	for {
// 		getResponce(*conn, fromStr, &resvCount)
// 	}
// }

// func flags2() (interval int, others int) {return}

// func localIP() string {
// 	return ""
// }

// func newDNS(addr string) (ip net.IP, from string) {
// 	return nil, ""
// }
