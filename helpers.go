package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"
)

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

func average(times []time.Duration) time.Duration {
	tottal := time.Duration(0)
	for _, val := range times {
		tottal += val
	}
	tottal = tottal / time.Duration(len(times))
	return tottal
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
		avg := average(stats.RTT)
		max := stats.RTT[len(stats.RTT)-1]
		mdev := stats.RTT[len(stats.RTT)/2]
		fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n", min.Seconds(), avg.Seconds(), max.Seconds(), mdev.Seconds())
	}
	os.Exit(0)
}

func flags() (destination string, intervals, size, ttl *int) {
	intervals = flag.Int("i", 1, "Wait interval between seding each packet.")
	size = flag.Int("s", 56, " Specifies the number of data bytes to be sent. The default is 56, which translates into 64 ICMP data bytes when combined with the 8 bytes of ICMP header data.")
	ttl = flag.Int("t", 50, "Set the IP Time to Live.")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("ping: usage error: Destination address required")
		os.Exit(1)
	}
	destination = flag.Arg(0)
	return
}
