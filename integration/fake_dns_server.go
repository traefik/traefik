package integration

import (
	"fmt"
	"net"
	"os"

	"github.com/containous/traefik/log"
	"github.com/miekg/dns"
)

type handler struct{}

// ServeDNS a fake DNS server
// Simplified version of the Challenge Test Server from Boulder
// https://github.com/letsencrypt/boulder/blob/a6597b9f120207eff192c3e4107a7e49972a0250/test/challtestsrv/dnsone.go#L40
func (s *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	fakeDNS := os.Getenv("DOCKER_HOST_IP")
	if fakeDNS == "" {
		fakeDNS = "127.0.0.1"
	}
	for _, q := range r.Question {
		log.Printf("Query -- [%s] %s", q.Name, dns.TypeToString[q.Qtype])

		switch q.Qtype {
		case dns.TypeA:
			record := new(dns.A)
			record.Hdr = dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    0,
			}
			record.A = net.ParseIP(fakeDNS)

			m.Answer = append(m.Answer, record)
		case dns.TypeCAA:
			addCAARecord := true

			var value string
			switch q.Name {
			case "bad-caa-reserved.com.":
				value = "sad-hacker-ca.invalid"
			case "good-caa-reserved.com.":
				value = "happy-hacker-ca.invalid"
			case "accounturi.good-caa-reserved.com.":
				uri := os.Getenv("ACCOUNT_URI")
				value = fmt.Sprintf("happy-hacker-ca.invalid; accounturi=%s", uri)
			case "recheck.good-caa-reserved.com.":
				// Allow issuance when we're running in the past
				// (under FAKECLOCK), otherwise deny issuance.
				if os.Getenv("FAKECLOCK") != "" {
					value = "happy-hacker-ca.invalid"
				} else {
					value = "sad-hacker-ca.invalid"
				}
			case "dns-01-only.good-caa-reserved.com.":
				value = "happy-hacker-ca.invalid; validationmethods=dns-01"
			case "http-01-only.good-caa-reserved.com.":
				value = "happy-hacker-ca.invalid; validationmethods=http-01"
			case "dns-01-or-http-01.good-caa-reserved.com.":
				value = "happy-hacker-ca.invalid; validationmethods=dns-01,http-01"
			default:
				addCAARecord = false
			}
			if addCAARecord {
				record := new(dns.CAA)
				record.Hdr = dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeCAA,
					Class:  dns.ClassINET,
					Ttl:    0,
				}
				record.Tag = "issue"
				record.Value = value
				m.Answer = append(m.Answer, record)
			}
		}
	}

	auth := new(dns.SOA)
	auth.Hdr = dns.RR_Header{Name: "boulder.invalid.", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 0}
	auth.Ns = "ns.boulder.invalid."
	auth.Mbox = "master.boulder.invalid."
	auth.Serial = 1
	auth.Refresh = 1
	auth.Retry = 1
	auth.Expire = 1
	auth.Minttl = 1
	m.Ns = append(m.Ns, auth)

	w.WriteMsg(m)
}

func startFakeDNSServer() *dns.Server {
	srv := &dns.Server{
		Addr:    ":5053",
		Net:     "udp",
		Handler: &handler{},
	}

	go func() {
		log.Infof("Start a fake DNS server.")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp listener %v", err)
		}
	}()

	return srv
}
