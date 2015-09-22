package main

import (
	"fmt"

	"github.com/alecthomas/kingpin"
)

var (
	debug   = kingpin.Flag("debug", "Enable debug mode.").Bool()
	timeout = kingpin.Flag("timeout", "Timeout waiting for ping.").OverrideDefaultFromEnvar("PING_TIMEOUT").Required().Short('t').Duration()
	ip      = kingpin.Arg("ip", "IP address to ping.").Required().IP()
	count   = kingpin.Arg("count", "Number of packets to send").Int()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	fmt.Printf("Would ping: %s with timeout %s and count %d", *ip, *timeout, *count)
}
