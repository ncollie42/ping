package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	"golang.org/x/net/icmp"
)

var gPing *ping

func init() {
	runtime := runtime.GOOS

	if runtime != "linux" && runtime != "darwin" {
		log.Fatal(runtime, " is not suported")
	}
	gPing = new(ping)
	flag.StringVar(&gPing.targetIP, "dest", "N/A", "dns name or ipv4 address")
	flag.IntVar(&gPing.interval, "i", 1, "Wait interval between seding each packet.")

	flag.Parse()
	if gPing.targetIP == "N/A" {
		log.Fatal("ping: usage error: Destination address required\n\tUse -dest <destination>") //change to not log
	}
}

func main() {
	var err error
	gPing.conn, err = icmp.ListenPacket("ip4:icmp", "127.0.0.1")
	if err != nil {
		log.Fatal(err)
	}
	defer gPing.conn.Close()

	fmt.Println(gPing)
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
