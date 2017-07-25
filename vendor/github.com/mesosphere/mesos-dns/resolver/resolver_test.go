package resolver

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	. "github.com/mesosphere/mesos-dns/dnstest"
	"github.com/mesosphere/mesos-dns/logging"
	"github.com/mesosphere/mesos-dns/records"
	"github.com/mesosphere/mesos-dns/records/labels"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/miekg/dns"
)

func init() {
	logging.VerboseFlag = false
	logging.SetupLogs()
}

// dig @127.0.0.1 -p 8053 "bob.*.mesos" ANY
func TestCleanWild(t *testing.T) {
	dom := "bob.*.mesos"

	stripped := cleanWild(dom)

	if stripped != "bob.mesos" {
		t.Error("not stripping domain")
	}
}

func TestShuffleAnswers(t *testing.T) {
	var res Resolver

	m := new(dns.Msg)

	for i := 0; i < 10; i++ {
		name := "10.0.0." + strconv.Itoa(i)
		rr, err := res.formatA("blah.com", name)
		if err != nil {
			t.Error(err)
		}
		m.Answer = append(m.Answer, rr)
	}

	n := new(dns.Msg)
	c := make([]dns.RR, len(m.Answer))
	copy(c, m.Answer)
	n.Answer = c

	rng := rand.New(rand.NewSource(0))
	_ = shuffleAnswers(rng, m.Answer)

	sflag := false
	// 10! chance of failing here
	for i := 0; i < 10; i++ {
		if n.Answer[i] != m.Answer[i] {
			sflag = true
			break
		}
	}

	if !sflag {
		t.Error("not shuffling")
	}
}

func TestHandlers(t *testing.T) {
	if err := runHandlers(); err != nil {
		t.Error(err)
	}
}

func BenchmarkHandlers(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := runHandlers(); err != nil {
			b.Error(err)
		}
	}
}

func runHandlers() error {
	res, err := fakeDNS()
	if err != nil {
		return err
	}
	res.fwd = func(m *dns.Msg, net string) (*dns.Msg, error) {
		rr1, err := res.formatA("google.com.", "1.1.1.1")
		if err != nil {
			return nil, err
		}
		rr2, err := res.formatA("google.com.", "2.2.2.2")
		if err != nil {
			return nil, err
		}
		msg := &dns.Msg{Answer: []dns.RR{rr1, rr2}}
		msg.SetReply(m)
		return msg, nil
	}

	for i, tt := range []struct {
		dns.HandlerFunc
		*dns.Msg
	}{
		{
			res.HandleMesos,
			Message(
				Question("chronos.marathon.mesos.", dns.TypeA),
				Header(true, dns.RcodeSuccess),
				Answers(
					A(RRHeader("chronos.marathon.mesos.", dns.TypeA, 60),
						net.ParseIP("1.2.3.11")))),
		},
		{ // case insensitive
			res.HandleMesos,
			Message(
				Question("cHrOnOs.MARATHON.mesoS.", dns.TypeA),
				Header(true, dns.RcodeSuccess),
				Answers(
					A(RRHeader("chronos.marathon.mesos.", dns.TypeA, 60),
						net.ParseIP("1.2.3.11")))),
		},
		{
			res.HandleMesos,
			Message(
				Question("_liquor-store._tcp.marathon.mesos.", dns.TypeSRV),
				Header(true, dns.RcodeSuccess),
				Answers(
					SRV(RRHeader("_liquor-store._tcp.marathon.mesos.", dns.TypeSRV, 60),
						"liquor-store-4dfjd-0.marathon.mesos.", 443, 0, 0),
					SRV(RRHeader("_liquor-store._tcp.marathon.mesos.", dns.TypeSRV, 60),
						"liquor-store-zasmd-1.marathon.mesos.", 80, 0, 0),
					SRV(RRHeader("_liquor-store._tcp.marathon.mesos.", dns.TypeSRV, 60),
						"liquor-store-zasmd-1.marathon.mesos.", 443, 0, 0),
					SRV(RRHeader("_liquor-store._tcp.marathon.mesos.", dns.TypeSRV, 60),
						"liquor-store-4dfjd-0.marathon.mesos.", 80, 0, 0)),
				Extras(
					A(RRHeader("liquor-store-4dfjd-0.marathon.mesos.", dns.TypeA, 60),
						net.ParseIP("10.3.0.1")),
					A(RRHeader("liquor-store-zasmd-1.marathon.mesos.", dns.TypeA, 60),
						net.ParseIP("10.3.0.2")))),
		},
		{
			res.HandleMesos,
			Message(
				Question("_car-store._udp.marathon.mesos.", dns.TypeSRV),
				Header(true, dns.RcodeSuccess),
				Answers(
					SRV(RRHeader("_car-store._udp.marathon.mesos.", dns.TypeSRV, 60),
						"car-store-zinaz-0.marathon.slave.mesos.", 31365, 0, 0),
					SRV(RRHeader("_car-store._udp.marathon.mesos.", dns.TypeSRV, 60),
						"car-store-zinaz-0.marathon.slave.mesos.", 31364, 0, 0)),
				Extras(
					A(RRHeader("car-store-zinaz-0.marathon.slave.mesos.", dns.TypeA, 60),
						net.ParseIP("1.2.3.11")))),
		},
		{
			res.HandleMesos,
			Message(
				Question("_car-store._udp.marathon.mesos.", dns.TypeA),
				Header(true, dns.RcodeSuccess),
				NSs(
					SOA(RRHeader("_car-store._udp.marathon.mesos.", dns.TypeSOA, 60),
						"ns1.mesos", "root.ns1.mesos", 60))),
		},
		{
			res.HandleMesos,
			Message(
				Question("non-existing.mesos.", dns.TypeSOA),
				Header(true, dns.RcodeSuccess),
				NSs(
					SOA(RRHeader("non-existing.mesos.", dns.TypeSOA, 60),
						"ns1.mesos", "root.ns1.mesos", 60))),
		},
		{
			res.HandleMesos,
			Message(
				Question("non-existing.mesos.", dns.TypeNS),
				Header(true, dns.RcodeSuccess),
				NSs(
					NS(RRHeader("non-existing.mesos.", dns.TypeNS, 60), "ns1.mesos"))),
		},
		{
			res.HandleMesos,
			Message(
				Question("missing.mesos.", dns.TypeA),
				Header(true, dns.RcodeNameError),
				NSs(
					SOA(RRHeader("missing.mesos.", dns.TypeSOA, 60),
						"ns1.mesos", "root.ns1.mesos", 60))),
		},
		{
			res.HandleMesos,
			Message(
				Question("chronos.marathon.mesos.", dns.TypeAAAA),
				Header(true, dns.RcodeSuccess),
				NSs(
					SOA(RRHeader("chronos.marathon.mesos.", dns.TypeSOA, 60),
						"ns1.mesos", "root.ns1.mesos", 60))),
		},
		{
			res.HandleMesos,
			Message(
				Question("missing.mesos.", dns.TypeAAAA),
				Header(true, dns.RcodeSuccess),
				NSs(
					SOA(RRHeader("missing.mesos.", dns.TypeSOA, 60),
						"ns1.mesos", "root.ns1.mesos", 60))),
		},
		{
			res.HandleNonMesos,
			Message(
				Question("google.com.", dns.TypeA),
				Header(false, dns.RcodeSuccess),
				Answers(
					A(RRHeader("google.com.", dns.TypeA, 60), net.ParseIP("1.1.1.1")),
					A(RRHeader("google.com.", dns.TypeA, 60), net.ParseIP("2.2.2.2")))),
		},
	} {
		var rw ResponseRecorder
		tt.HandlerFunc(&rw, tt.Msg)
		if got, want := rw.Msg, tt.Msg; !(Msg{got}).equivalent(Msg{want}) {
			return fmt.Errorf("Test #%d\n%v\n%s\n", i, pretty.Sprint(tt.Msg.Question), pretty.Compare(got, want))
		}
	}
	return nil
}

type Msg struct{ *dns.Msg }
type RRs []dns.RR

func (m Msg) equivalent(other Msg) bool {
	if m.Msg == nil || other.Msg == nil {
		return m.Msg == other.Msg
	}
	return m.MsgHdr == other.MsgHdr &&
		m.Compress == other.Compress &&
		reflect.DeepEqual(m.Question, other.Question) &&
		RRs(m.Ns).equivalent(RRs(other.Ns)) &&
		RRs(m.Answer).equivalent(RRs(other.Answer)) &&
		RRs(m.Extra).equivalent(RRs(other.Extra))
}

// equivalent RRs have the same records, but not necessarily in the same order
func (rr RRs) equivalent(other RRs) bool {
	if rr == nil || other == nil {
		return rr == nil && other == nil
	}
	type key struct {
		header dns.RR_Header
		text   string
	}

	rrhash := make(map[string]struct{}, len(rr))
	for i := range rr {
		var k key
		header := rr[i].Header()
		if header != nil {
			k.header = *header
		}
		k.text = rr[i].String()
		s := fmt.Sprintf("%+v", k)
		rrhash[s] = struct{}{}
	}

	for i := range other {
		var k key
		header := other[i].Header()
		if header != nil {
			k.header = *header
		}
		k.text = other[i].String()
		s := fmt.Sprintf("%+v", k)
		if _, ok := rrhash[s]; !ok {
			return false
		}
		delete(rrhash, s)
	}
	return len(rrhash) == 0
}

func TestHTTP(t *testing.T) {
	// setup DNS server (just http)
	res, err := fakeDNS()
	if err != nil {
		t.Fatal(err)
	}
	res.version = "0.1.1"

	res.configureHTTP()
	srv := httptest.NewServer(http.DefaultServeMux)
	defer srv.Close()

	for _, tt := range []struct {
		path      string
		code      int
		got, want interface{}
	}{
		{"/v1/version", http.StatusOK, map[string]interface{}{},
			map[string]interface{}{
				"Service": "Mesos-DNS",
				"URL":     "https://github.com/mesosphere/mesos-dns",
				"Version": "0.1.1",
			},
		},
		{"/v1/config", http.StatusOK, &records.Config{}, &res.config},
		{"/v1/services/_leader._tcp.mesos.", http.StatusOK, []interface{}{},
			[]interface{}{map[string]interface{}{
				"service": "_leader._tcp.mesos.",
				"host":    "leader.mesos.",
				"ip":      "1.2.3.4",
				"port":    "5050",
			}},
		},
		{"/v1/services/_myservice._tcp.mesos.", http.StatusOK, []interface{}{},
			[]interface{}{map[string]interface{}{
				"service": "",
				"host":    "",
				"ip":      "",
				"port":    "",
			}},
		},
		{"/v1/hosts/leader.mesos", http.StatusOK, []interface{}{},
			[]interface{}{map[string]interface{}{
				"host": "leader.mesos.",
				"ip":   "1.2.3.4",
			}},
		},
	} {
		if resp, err := http.Get(srv.URL + tt.path); err != nil {
			t.Error(err)
		} else if got, want := resp.StatusCode, tt.code; got != want {
			t.Errorf("GET %s: StatusCode: got %d, want %d", tt.path, got, want)
		} else if err := json.NewDecoder(resp.Body).Decode(&tt.got); err != nil {
			t.Error(err)
		} else if got, want := tt.got, tt.want; !reflect.DeepEqual(got, want) {
			t.Errorf("GET %s: Body:\ngot:  %#v\nwant: %#v", tt.path, got, want)
		} else {
			_ = resp.Body.Close()
		}
	}
}

func fakeDNS() (*Resolver, error) {
	config := records.NewConfig()
	config.Masters = []string{"144.76.157.37:5050"}
	config.RecurseOn = false
	config.IPSources = []string{"docker", "mesos", "host"}

	res := New("", config)
	res.rng.Seed(0) // for deterministic tests

	b, err := ioutil.ReadFile("../factories/fake.json")
	if err != nil {
		return nil, err
	}

	var sj state.State
	err = json.Unmarshal(b, &sj)
	if err != nil {
		return nil, err
	}

	spec := labels.RFC952
	err = res.rs.InsertState(sj, "mesos", "mesos-dns.mesos.", "127.0.0.1", res.config.Masters, res.config.IPSources, spec)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func onError(abort <-chan struct{}, errCh <-chan error, f func(error)) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		select {
		case <-abort:
		case e := <-errCh:
			if e != nil {
				defer close(ch)
				f(e)
			}
		}
	}()
	return ch
}

func TestMultiError(t *testing.T) {
	me := multiError(nil)
	me.Add()
	me.Add(nil)
	me.Add(multiError(nil))
	if !me.Nil() {
		t.Fatalf("Expected Nil() multiError instead of %q", me.Error())
	}

	me.Add(errors.New("abc"))
	me.Add(errors.New("123"))
	me.Add(multiError(nil))
	me.Add(multiError([]error{errors.New("456")}))
	me.Add(multiError{errors.New("789")})
	me.Add(errors.New("def"))

	const expected = "abc; 123; 456; 789; def"
	actual := me.Error()
	if expected != actual {
		t.Fatalf("expected %q instead of %q", expected, actual)
	}
}

func TestTruncate(t *testing.T) {
	tm := newTruncated()
	if !tm.Truncated {
		t.Fatal("Message not truncated")
	}
	if l := tm.Len(); l > 512 {
		t.Fatalf("Message to large: %d bytes", l)
	}
	tm.Answer = append(tm.Answer, genA(1)...)
	if l := tm.Len(); l < 512 {
		t.Fatalf("Message to small after adding answers: %d bytes", l)
	}
}

func BenchmarkTruncate(b *testing.B) {
	for n := 0; n < b.N; n++ {
		newTruncated()
	}
}

func newTruncated() *dns.Msg {
	m := Message(
		Question("example.com.", dns.TypeA),
		Header(false, dns.RcodeSuccess),
		Answers(genA(50)...))

	return truncate(m, true)
}

func genA(n int) []dns.RR {
	records := make([]dns.RR, n)
	ip := []byte{0, 0, 0, 0}
	for i := 0; i < n; i++ {
		binary.PutUvarint(ip, uint64(i))
		records[i] = A(RRHeader("example.com.", dns.TypeA, 60), ip)
	}
	return records
}
