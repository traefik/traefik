package main

import (
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
)

func query(dom string) {
	nameserver := "127.0.0.1:8053"

	qt := dns.TypeA
	qc := uint16(dns.ClassINET)

	c := new(dns.Client)
	c.Net = "udp"

	m := new(dns.Msg)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{
		Name:   dns.Fqdn(dom),
		Qtype:  qt,
		Qclass: qc,
	}

	_, _, err := c.Exchange(m, nameserver)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	start := time.Now()

	cnt := 10000

	for i := 0; i < cnt; i++ {
		go query("bob.mesos")
	}

	elapsed := time.Since(start)
	log.Printf("benching took %s", elapsed)
	log.Printf("doing %d/%v rps", cnt, elapsed)
}
