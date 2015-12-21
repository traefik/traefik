package main

import (
	"net"
	"net/http"
	"os/exec"
	"sync"
	"time"
    "github.com/miekg/dns"

	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
)

func (s *CNameSuite) TestSimpleConfiguration(c *check.C) {
    // Start traefik with a specific resolve.conf configuration
	cmd := exec.Command(traefikBinary, "fixtures/cname/cname.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

    //
    name := "notthere.localhost"
    url := "http://" + name + ":80"

    dns.HandleFunc(name, CnameServer)
    defer dns.HandleRemove(name)

    // Run our own DNS server to serve the cname record we want
    server, _, err := runLocalUDPServer("127.0.0.1:0")
    if err != nil {
        c.Fatalf("unable to run test server: %v", err)
    }
    defer server.Shutdown()

	time.Sleep(1000 * time.Millisecond)
	resp, err := http.Get(url)

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}


func CnameServer(w dns.ResponseWriter, req *dns.Msg) {
  m := new(dns.Msg)
  m.SetReply(req)

  m.Extra = make([]dns.RR, 1)
  m.Extra[0] = &dns.CNAME{Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 0}, Target: "test.localhost"}
  w.WriteMsg(m)

}

func runLocalUDPServer(laddr string) (*dns.Server, string, error) {
	pc, err := net.ListenPacket("udp", laddr)
	if err != nil {
		return nil, "", err
	}
	server := &dns.Server{PacketConn: pc, ReadTimeout: time.Hour, WriteTimeout: time.Hour}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	go func() {
		server.ActivateAndServe()
		pc.Close()
	}()

	waitLock.Lock()
	return server, pc.LocalAddr().String(), nil
}
