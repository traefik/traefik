package rfc2136

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/xenolf/lego/acme"
)

var (
	rfc2136TestDomain     = "123456789.www.example.com"
	rfc2136TestKeyAuth    = "123d=="
	rfc2136TestValue      = "Now36o-3BmlB623-0c1qCIUmgWVVmDJb88KGl24pqpo"
	rfc2136TestFqdn       = "_acme-challenge.123456789.www.example.com."
	rfc2136TestZone       = "example.com."
	rfc2136TestTTL        = 120
	rfc2136TestTsigKey    = "example.com."
	rfc2136TestTsigSecret = "IwBTJx9wrDp4Y1RyC3H0gA=="
)

var reqChan = make(chan *dns.Msg, 10)

func TestRFC2136CanaryLocalTestServer(t *testing.T) {
	acme.ClearFqdnCache()
	dns.HandleFunc("example.com.", serverHandlerHello)
	defer dns.HandleRemove("example.com.")

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeTXT)
	r, _, err := c.Exchange(m, addrstr)
	if err != nil || len(r.Extra) == 0 {
		t.Fatalf("Failed to communicate with test server: %v", err)
	}
	txt := r.Extra[0].(*dns.TXT).Txt[0]
	if txt != "Hello world" {
		t.Error("Expected test server to return 'Hello world' but got: ", txt)
	}
}

func TestRFC2136ServerSuccess(t *testing.T) {
	acme.ClearFqdnCache()
	dns.HandleFunc(rfc2136TestZone, serverHandlerReturnSuccess)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	provider, err := NewDNSProviderCredentials(addrstr, "", "", "")
	if err != nil {
		t.Fatalf("Expected NewDNSProviderCredentials() to return no error but the error was -> %v", err)
	}
	if err := provider.Present(rfc2136TestDomain, "", rfc2136TestKeyAuth); err != nil {
		t.Errorf("Expected Present() to return no error but the error was -> %v", err)
	}
}

func TestRFC2136ServerError(t *testing.T) {
	acme.ClearFqdnCache()
	dns.HandleFunc(rfc2136TestZone, serverHandlerReturnErr)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	provider, err := NewDNSProviderCredentials(addrstr, "", "", "")
	if err != nil {
		t.Fatalf("Expected NewDNSProviderCredentials() to return no error but the error was -> %v", err)
	}
	if err := provider.Present(rfc2136TestDomain, "", rfc2136TestKeyAuth); err == nil {
		t.Errorf("Expected Present() to return an error but it did not.")
	} else if !strings.Contains(err.Error(), "NOTZONE") {
		t.Errorf("Expected Present() to return an error with the 'NOTZONE' rcode string but it did not.")
	}
}

func TestRFC2136TsigClient(t *testing.T) {
	acme.ClearFqdnCache()
	dns.HandleFunc(rfc2136TestZone, serverHandlerReturnSuccess)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", true)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	provider, err := NewDNSProviderCredentials(addrstr, "", rfc2136TestTsigKey, rfc2136TestTsigSecret)
	if err != nil {
		t.Fatalf("Expected NewDNSProviderCredentials() to return no error but the error was -> %v", err)
	}
	if err := provider.Present(rfc2136TestDomain, "", rfc2136TestKeyAuth); err != nil {
		t.Errorf("Expected Present() to return no error but the error was -> %v", err)
	}
}

func TestRFC2136ValidUpdatePacket(t *testing.T) {
	acme.ClearFqdnCache()
	dns.HandleFunc(rfc2136TestZone, serverHandlerPassBackRequest)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	txtRR, _ := dns.NewRR(fmt.Sprintf("%s %d IN TXT %s", rfc2136TestFqdn, rfc2136TestTTL, rfc2136TestValue))
	rrs := []dns.RR{txtRR}
	m := new(dns.Msg)
	m.SetUpdate(rfc2136TestZone)
	m.RemoveRRset(rrs)
	m.Insert(rrs)
	expectstr := m.String()
	expect, err := m.Pack()
	if err != nil {
		t.Fatalf("Error packing expect msg: %v", err)
	}

	provider, err := NewDNSProviderCredentials(addrstr, "", "", "")
	if err != nil {
		t.Fatalf("Expected NewDNSProviderCredentials() to return no error but the error was -> %v", err)
	}

	if err := provider.Present(rfc2136TestDomain, "", "1234d=="); err != nil {
		t.Errorf("Expected Present() to return no error but the error was -> %v", err)
	}

	rcvMsg := <-reqChan
	rcvMsg.Id = m.Id
	actual, err := rcvMsg.Pack()
	if err != nil {
		t.Fatalf("Error packing actual msg: %v", err)
	}

	if !bytes.Equal(actual, expect) {
		tmp := new(dns.Msg)
		if err := tmp.Unpack(actual); err != nil {
			t.Fatalf("Error unpacking actual msg: %v", err)
		}
		t.Errorf("Expected msg:\n%s", expectstr)
		t.Errorf("Actual msg:\n%v", tmp)
	}
}

func runLocalDNSTestServer(listenAddr string, tsig bool) (*dns.Server, string, error) {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return nil, "", err
	}
	server := &dns.Server{PacketConn: pc, ReadTimeout: time.Hour, WriteTimeout: time.Hour}
	if tsig {
		server.TsigSecret = map[string]string{rfc2136TestTsigKey: rfc2136TestTsigSecret}
	}

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

func serverHandlerHello(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Extra = make([]dns.RR, 1)
	m.Extra[0] = &dns.TXT{
		Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
		Txt: []string{"Hello world"},
	}
	w.WriteMsg(m)
}

func serverHandlerReturnSuccess(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	if req.Opcode == dns.OpcodeQuery && req.Question[0].Qtype == dns.TypeSOA && req.Question[0].Qclass == dns.ClassINET {
		// Return SOA to appease findZoneByFqdn()
		soaRR, _ := dns.NewRR(fmt.Sprintf("%s %d IN SOA ns1.%s admin.%s 2016022801 28800 7200 2419200 1200", rfc2136TestZone, rfc2136TestTTL, rfc2136TestZone, rfc2136TestZone))
		m.Answer = []dns.RR{soaRR}
	}

	if t := req.IsTsig(); t != nil {
		if w.TsigStatus() == nil {
			// Validated
			m.SetTsig(rfc2136TestZone, dns.HmacMD5, 300, time.Now().Unix())
		}
	}

	w.WriteMsg(m)
}

func serverHandlerReturnErr(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeNotZone)
	w.WriteMsg(m)
}

func serverHandlerPassBackRequest(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	if req.Opcode == dns.OpcodeQuery && req.Question[0].Qtype == dns.TypeSOA && req.Question[0].Qclass == dns.ClassINET {
		// Return SOA to appease findZoneByFqdn()
		soaRR, _ := dns.NewRR(fmt.Sprintf("%s %d IN SOA ns1.%s admin.%s 2016022801 28800 7200 2419200 1200", rfc2136TestZone, rfc2136TestTTL, rfc2136TestZone, rfc2136TestZone))
		m.Answer = []dns.RR{soaRR}
	}

	if t := req.IsTsig(); t != nil {
		if w.TsigStatus() == nil {
			// Validated
			m.SetTsig(rfc2136TestZone, dns.HmacMD5, 300, time.Now().Unix())
		}
	}

	w.WriteMsg(m)
	if req.Opcode != dns.OpcodeQuery || req.Question[0].Qtype != dns.TypeSOA || req.Question[0].Qclass != dns.ClassINET {
		// Only talk back when it is not the SOA RR.
		reqChan <- req
	}
}
