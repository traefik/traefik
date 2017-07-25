package records

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/mesosphere/mesos-dns/logging"
	"github.com/mesosphere/mesos-dns/records/labels"
	"github.com/mesosphere/mesos-dns/records/state"
)

func init() {
	logging.VerboseFlag = false
	logging.VeryVerboseFlag = false
	logging.SetupLogs()
}

func (rg *RecordGenerator) exists(name, host string, kind rrsKind) bool {
	rrs := kind.rrs(rg)
	if val, ok := rrs[name]; ok {
		_, ok := val[host]
		return ok
	}
	return false
}

func TestMasterRecord(t *testing.T) {
	// masterRecord(domain string, masters []string, leader string)
	type expectedRR struct {
		name string
		host string
		kind rrsKind
	}
	tt := []struct {
		domain  string
		masters []string
		leader  string
		expect  []expectedRR
	}{
		{"foo.com", nil, "", nil},
		{"foo.com", nil, "@", nil},
		{"foo.com", nil, "1@", nil},
		{"foo.com", nil, "@2", nil},
		{"foo.com", nil, "3@4", nil},
		{"foo.com", nil, "5@6:7",
			[]expectedRR{
				{"leader.foo.com.", "6", A},
				{"master.foo.com.", "6", A},
				{"master0.foo.com.", "6", A},
				{"_leader._tcp.foo.com.", "leader.foo.com.:7", SRV},
				{"_leader._udp.foo.com.", "leader.foo.com.:7", SRV},
			}},
		// single master: leader and fallback
		{"foo.com", []string{"6:7"}, "5@6:7",
			[]expectedRR{
				{"leader.foo.com.", "6", A},
				{"master.foo.com.", "6", A},
				{"master0.foo.com.", "6", A},
				{"_leader._tcp.foo.com.", "leader.foo.com.:7", SRV},
				{"_leader._udp.foo.com.", "leader.foo.com.:7", SRV},
			}},
		// leader not in fallback list
		{"foo.com", []string{"8:9"}, "5@6:7",
			[]expectedRR{
				{"leader.foo.com.", "6", A},
				{"master.foo.com.", "6", A},
				{"master.foo.com.", "8", A},
				{"master1.foo.com.", "6", A},
				{"master0.foo.com.", "8", A},
				{"_leader._tcp.foo.com.", "leader.foo.com.:7", SRV},
				{"_leader._udp.foo.com.", "leader.foo.com.:7", SRV},
			}},
		// duplicate fallback masters, leader not in fallback list
		{"foo.com", []string{"8:9", "8:9"}, "5@6:7",
			[]expectedRR{
				{"leader.foo.com.", "6", A},
				{"master.foo.com.", "6", A},
				{"master.foo.com.", "8", A},
				{"master1.foo.com.", "6", A},
				{"master0.foo.com.", "8", A},
				{"_leader._tcp.foo.com.", "leader.foo.com.:7", SRV},
				{"_leader._udp.foo.com.", "leader.foo.com.:7", SRV},
			}},
		// leader that's also listed in the fallback list (at the end)
		{"foo.com", []string{"8:9", "6:7"}, "5@6:7",
			[]expectedRR{
				{"leader.foo.com.", "6", A},
				{"master.foo.com.", "6", A},
				{"master.foo.com.", "8", A},
				{"master1.foo.com.", "6", A},
				{"master0.foo.com.", "8", A},
				{"_leader._tcp.foo.com.", "leader.foo.com.:7", SRV},
				{"_leader._udp.foo.com.", "leader.foo.com.:7", SRV},
			}},
		// duplicate leading masters in the fallback list
		{"foo.com", []string{"8:9", "6:7", "6:7"}, "5@6:7",
			[]expectedRR{
				{"leader.foo.com.", "6", A},
				{"master.foo.com.", "6", A},
				{"master.foo.com.", "8", A},
				{"master1.foo.com.", "6", A},
				{"master0.foo.com.", "8", A},
				{"_leader._tcp.foo.com.", "leader.foo.com.:7", SRV},
				{"_leader._udp.foo.com.", "leader.foo.com.:7", SRV},
			}},
		// leader that's also listed in the fallback list (in the middle)
		{"foo.com", []string{"8:9", "6:7", "bob:0"}, "5@6:7",
			[]expectedRR{
				{"leader.foo.com.", "6", A},
				{"master.foo.com.", "6", A},
				{"master.foo.com.", "8", A},
				{"master.foo.com.", "bob", A},
				{"master0.foo.com.", "8", A},
				{"master1.foo.com.", "6", A},
				{"master2.foo.com.", "bob", A},
				{"_leader._tcp.foo.com.", "leader.foo.com.:7", SRV},
				{"_leader._udp.foo.com.", "leader.foo.com.:7", SRV},
			}},
	}
	for i, tc := range tt {
		rg := &RecordGenerator{}
		rg.As = make(rrs)
		rg.SRVs = make(rrs)
		t.Logf("test case %d", i+1)
		rg.masterRecord(tc.domain, tc.masters, tc.leader)
		if tc.expect == nil {
			if len(rg.As) > 0 {
				t.Fatalf("test case %d: unexpected As: %v", i+1, rg.As)
			}
			if len(rg.SRVs) > 0 {
				t.Fatalf("test case %d: unexpected SRVs: %v", i+1, rg.SRVs)
			}
		}
		expectedA := make(rrs)
		expectedSRV := make(rrs)
		for _, e := range tc.expect {
			found := rg.exists(e.name, e.host, e.kind)
			if !found {
				t.Fatalf("test case %d: missing expected record: name=%q host=%q kind=%s, As=%v", i+1, e.name, e.host, e.kind, rg.As)
			}
			switch e.kind {
			case A:
				expectedA.add(e.name, e.host)
			case SRV:
				expectedSRV.add(e.name, e.host)
			default:
				t.Fatalf("unexpected kind %q", e.kind)
			}
		}
		if !reflect.DeepEqual(rg.As, expectedA) {
			t.Fatalf("test case %d: expected As of %v instead of %v", i+1, expectedA, rg.As)
		}
		if !reflect.DeepEqual(rg.SRVs, expectedSRV) {
			t.Fatalf("test case %d: expected SRVs of %v instead of %v", i+1, expectedSRV, rg.SRVs)
		}
	}
}

func TestLeaderIP(t *testing.T) {
	l := "master@144.76.157.37:5050"

	ip := leaderIP(l)

	if ip != "144.76.157.37" {
		t.Error("not parsing ip")
	}
}

func testRecordGenerator(t *testing.T, spec labels.Func, ipSources []string) RecordGenerator {
	var sj state.State

	b, err := ioutil.ReadFile("../factories/fake.json")
	if err != nil {
		t.Fatal(err)
	} else if err = json.Unmarshal(b, &sj); err != nil {
		t.Fatal(err)
	}

	sj.Leader = "master@144.76.157.37:5050"
	masters := []string{"144.76.157.37:5050"}

	var rg RecordGenerator
	if err := rg.InsertState(sj, "mesos", "mesos-dns.mesos.", "127.0.0.1", masters, ipSources, spec); err != nil {
		t.Fatal(err)
	}

	return rg
}

// ensure we are parsing what we think we are
func TestInsertState(t *testing.T) {
	rg := testRecordGenerator(t, labels.RFC952, []string{"docker", "mesos", "host"})
	rgDocker := testRecordGenerator(t, labels.RFC952, []string{"docker", "host"})
	rgMesos := testRecordGenerator(t, labels.RFC952, []string{"mesos", "host"})
	rgSlave := testRecordGenerator(t, labels.RFC952, []string{"host"})

	for i, tt := range []struct {
		rrs  rrs
		name string
		want []string
	}{
		{rg.As, "big-dog.marathon.mesos.", []string{"10.3.0.1"}},
		{rg.As, "liquor-store.marathon.mesos.", []string{"10.3.0.1", "10.3.0.2"}},
		{rg.As, "liquor-store.marathon.slave.mesos.", []string{"1.2.3.11", "1.2.3.12"}},
		{rg.As, "car-store.marathon.slave.mesos.", []string{"1.2.3.11"}},
		{rg.As, "nginx.marathon.mesos.", []string{"10.3.0.3"}},
		{rg.As, "poseidon.marathon.mesos.", nil},
		{rg.As, "poseidon.marathon.slave.mesos.", nil},
		{rg.As, "master.mesos.", []string{"144.76.157.37"}},
		{rg.As, "master0.mesos.", []string{"144.76.157.37"}},
		{rg.As, "leader.mesos.", []string{"144.76.157.37"}},
		{rg.As, "slave.mesos.", []string{"1.2.3.10", "1.2.3.11", "1.2.3.12"}},
		{rg.As, "some-box.chronoswithaspaceandmixe.mesos.", []string{"1.2.3.11"}}, // ensure we translate the framework name as well
		{rg.As, "marathon.mesos.", []string{"1.2.3.11"}},
		{rg.SRVs, "_big-dog._tcp.marathon.mesos.", []string{
			"big-dog-4dfjd-0.marathon.mesos.:80",
			"big-dog-4dfjd-0.marathon.mesos.:443",
		}},
		{rg.SRVs, "_poseidon._tcp.marathon.mesos.", nil},
		{rg.SRVs, "_leader._tcp.mesos.", []string{"leader.mesos.:5050"}},
		{rg.SRVs, "_liquor-store._tcp.marathon.mesos.", []string{
			"liquor-store-4dfjd-0.marathon.mesos.:80",
			"liquor-store-4dfjd-0.marathon.mesos.:443",
			"liquor-store-zasmd-1.marathon.mesos.:80",
			"liquor-store-zasmd-1.marathon.mesos.:443",
		}},
		{rg.SRVs, "_liquor-store._udp.marathon.mesos.", nil},
		{rg.SRVs, "_liquor-store.marathon.mesos.", nil},
		{rg.SRVs, "_https._liquor-store._tcp.marathon.mesos.", []string{
			"liquor-store-4dfjd-0.marathon.mesos.:443",
			"liquor-store-zasmd-1.marathon.mesos.:443",
		}},
		{rg.SRVs, "_http._liquor-store._tcp.marathon.mesos.", []string{
			"liquor-store-4dfjd-0.marathon.mesos.:80",
			"liquor-store-zasmd-1.marathon.mesos.:80",
		}},
		{rg.SRVs, "_car-store._tcp.marathon.mesos.", []string{
			"car-store-zinaz-0.marathon.slave.mesos.:31364",
			"car-store-zinaz-0.marathon.slave.mesos.:31365",
		}},
		{rg.SRVs, "_car-store._udp.marathon.mesos.", []string{
			"car-store-zinaz-0.marathon.slave.mesos.:31364",
			"car-store-zinaz-0.marathon.slave.mesos.:31365",
		}},
		{rg.SRVs, "_slave._tcp.mesos.", []string{"slave.mesos.:5051"}},
		{rg.SRVs, "_framework._tcp.marathon.mesos.", []string{"marathon.mesos.:25501"}},

		{rgSlave.As, "liquor-store.marathon.mesos.", []string{"1.2.3.11", "1.2.3.12"}},
		{rgSlave.As, "liquor-store.marathon.slave.mesos.", []string{"1.2.3.11", "1.2.3.12"}},
		{rgSlave.As, "nginx.marathon.mesos.", []string{"1.2.3.11"}},
		{rgSlave.As, "car-store.marathon.slave.mesos.", []string{"1.2.3.11"}},

		{rgMesos.As, "liquor-store.marathon.mesos.", []string{"1.2.3.11", "1.2.3.12"}},
		{rgMesos.As, "liquor-store.marathon.slave.mesos.", []string{"1.2.3.11", "1.2.3.12"}},
		{rgMesos.As, "nginx.marathon.mesos.", []string{"10.3.0.3"}},
		{rgMesos.As, "car-store.marathon.slave.mesos.", []string{"1.2.3.11"}},

		{rgDocker.As, "liquor-store.marathon.mesos.", []string{"10.3.0.1", "10.3.0.2"}},
		{rgDocker.As, "liquor-store.marathon.slave.mesos.", []string{"1.2.3.11", "1.2.3.12"}},
		{rgDocker.As, "nginx.marathon.mesos.", []string{"1.2.3.11"}},
		{rgDocker.As, "car-store.marathon.slave.mesos.", []string{"1.2.3.11"}},
	} {
		// convert want into a map[string]struct{} (string set) for simpler comparison
		// via reflect.DeepEqual
		want := map[string]struct{}{}
		for _, x := range tt.want {
			want[x] = struct{}{}
		}
		if got := tt.rrs[tt.name]; !reflect.DeepEqual(got, want) {
			if len(got) == 0 && len(want) == 0 {
				continue
			}
			t.Errorf("test #%d: %q: got: %q, want: %q", i, tt.name, got, want)
		}
	}
}

// ensure we only generate one A record for each host
func TestNTasks(t *testing.T) {
	rg := &RecordGenerator{}
	rg.As = make(rrs)

	rg.insertRR("blah.mesos", "10.0.0.1", A)
	rg.insertRR("blah.mesos", "10.0.0.1", A)
	rg.insertRR("blah.mesos", "10.0.0.2", A)

	k, _ := rg.As["blah.mesos"]

	if len(k) != 2 {
		t.Error("should only have 2 A records")
	}
}

func TestHashString(t *testing.T) {
	val := hashString("test")
	if len(val) != 5 {
		t.Fatal("Hash length not 5")
	}
	if val != "iffe9" {
		t.Fatal("hashString(test) != iffe9")
	}
}
func TestHashStringCollisions(t *testing.T) {
	if testing.Short() {
		t.Skip("Quickcheck - skipping for short mode.")
	}

	// Quickcheck config with seed from random.org
	// for deterministic behaviour
	quickConfig := &quick.Config{
		MaxCount: 1e7,
		Rand:     rand.New(rand.NewSource(34553613)),
	}
	fn := func(a, b string) bool { return (a == b) || (hashString(a) != hashString(b)) }
	if err := quick.Check(fn, quickConfig); err != nil {
		t.Fatal("hashString collision encountered ", err)
	}
}

func TestTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - skipping for short mode.")
	}

	sleepForeverHandler := func(w http.ResponseWriter, req *http.Request) {
		req.Close = true
		notify := w.(http.CloseNotifier).CloseNotify()
		<-notify
	}
	server := httptest.NewServer(http.HandlerFunc(sleepForeverHandler))
	defer server.Close()

	rg := NewRecordGenerator(500 * time.Millisecond)
	host, port, err := net.SplitHostPort(server.Listener.Addr().String())
	_, err = rg.loadFromMaster(host, port)
	if err == nil {
		t.Fatal("Expect error because of timeout handler")
	}
	urlErr, ok := (err).(*url.Error)
	if !ok {
		t.Fatalf("Expected url.Error, instead: %#v", urlErr)
	}
	netErr, ok := urlErr.Err.(net.Error)
	if !ok {
		t.Fatalf("Expected net.Error, instead: %#v", netErr)
	}
	if !netErr.Timeout() {
		t.Errorf("Did not receive a timeout, instead: %#v", err)
	}
}
