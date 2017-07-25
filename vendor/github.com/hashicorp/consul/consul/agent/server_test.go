package agent_test

import (
	"net"
	"testing"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/serf/serf"
)

func TestServer_Key_params(t *testing.T) {
	ipv4a := net.ParseIP("127.0.0.1")
	ipv4b := net.ParseIP("1.2.3.4")

	tests := []struct {
		name  string
		sd1   *agent.Server
		sd2   *agent.Server
		equal bool
	}{
		{
			name: "Addr inequality",
			sd1: &agent.Server{
				Name:       "s1",
				Datacenter: "dc1",
				Port:       8300,
				Addr:       &net.IPAddr{IP: ipv4a},
			},
			sd2: &agent.Server{
				Name:       "s1",
				Datacenter: "dc1",
				Port:       8300,
				Addr:       &net.IPAddr{IP: ipv4b},
			},
			equal: true,
		},
	}

	for _, test := range tests {
		if test.sd1.Key().Equal(test.sd2.Key()) != test.equal {
			t.Errorf("Expected a %v result from test %s", test.equal, test.name)
		}

		// Test Key to make sure it actually works as a key
		m := make(map[agent.Key]bool)
		m[*test.sd1.Key()] = true
		if _, found := m[*test.sd2.Key()]; found != test.equal {
			t.Errorf("Expected a %v result from map test %s", test.equal, test.name)
		}
	}
}

func TestIsConsulServer(t *testing.T) {
	m := serf.Member{
		Name: "foo",
		Addr: net.IP([]byte{127, 0, 0, 1}),
		Tags: map[string]string{
			"role":          "consul",
			"id":            "asdf",
			"dc":            "east-aws",
			"port":          "10000",
			"build":         "0.8.0",
			"wan_join_port": "1234",
			"vsn":           "1",
			"expect":        "3",
			"raft_vsn":      "3",
		},
		Status: serf.StatusLeft,
	}
	ok, parts := agent.IsConsulServer(m)
	if !ok || parts.Datacenter != "east-aws" || parts.Port != 10000 {
		t.Fatalf("bad: %v %v", ok, parts)
	}
	if parts.Name != "foo" {
		t.Fatalf("bad: %v", parts)
	}
	if parts.ID != "asdf" {
		t.Fatalf("bad: %v", parts.ID)
	}
	if parts.Bootstrap {
		t.Fatalf("unexpected bootstrap")
	}
	if parts.Expect != 3 {
		t.Fatalf("bad: %v", parts.Expect)
	}
	if parts.Port != 10000 {
		t.Fatalf("bad: %v", parts.Port)
	}
	if parts.WanJoinPort != 1234 {
		t.Fatalf("bad: %v", parts.WanJoinPort)
	}
	if parts.RaftVersion != 3 {
		t.Fatalf("bad: %v", parts.RaftVersion)
	}
	if parts.Status != serf.StatusLeft {
		t.Fatalf("bad: %v", parts.Status)
	}
	m.Tags["bootstrap"] = "1"
	m.Tags["disabled"] = "1"
	ok, parts = agent.IsConsulServer(m)
	if !ok {
		t.Fatalf("expected a valid consul server")
	}
	if !parts.Bootstrap {
		t.Fatalf("expected bootstrap")
	}
	if parts.Addr.String() != "127.0.0.1:10000" {
		t.Fatalf("bad addr: %v", parts.Addr)
	}
	if parts.Version != 1 {
		t.Fatalf("bad: %v", parts)
	}
	m.Tags["expect"] = "3"
	delete(m.Tags, "bootstrap")
	delete(m.Tags, "disabled")
	ok, parts = agent.IsConsulServer(m)
	if !ok || parts.Expect != 3 {
		t.Fatalf("bad: %v", parts.Expect)
	}
	if parts.Bootstrap {
		t.Fatalf("unexpected bootstrap")
	}

	delete(m.Tags, "role")
	ok, parts = agent.IsConsulServer(m)
	if ok {
		t.Fatalf("unexpected ok server")
	}
}

func TestIsConsulServer_Optional(t *testing.T) {
	m := serf.Member{
		Name: "foo",
		Addr: net.IP([]byte{127, 0, 0, 1}),
		Tags: map[string]string{
			"role":  "consul",
			"id":    "asdf",
			"dc":    "east-aws",
			"port":  "10000",
			"vsn":   "1",
			"build": "0.8.0",
			// wan_join_port, raft_vsn, and expect are optional and
			// should default to zero.
		},
	}
	ok, parts := agent.IsConsulServer(m)
	if !ok || parts.Datacenter != "east-aws" || parts.Port != 10000 {
		t.Fatalf("bad: %v %v", ok, parts)
	}
	if parts.Name != "foo" {
		t.Fatalf("bad: %v", parts)
	}
	if parts.ID != "asdf" {
		t.Fatalf("bad: %v", parts.ID)
	}
	if parts.Bootstrap {
		t.Fatalf("unexpected bootstrap")
	}
	if parts.Expect != 0 {
		t.Fatalf("bad: %v", parts.Expect)
	}
	if parts.Port != 10000 {
		t.Fatalf("bad: %v", parts.Port)
	}
	if parts.WanJoinPort != 0 {
		t.Fatalf("bad: %v", parts.WanJoinPort)
	}
	if parts.RaftVersion != 0 {
		t.Fatalf("bad: %v", parts.RaftVersion)
	}
	m.Tags["bootstrap"] = "1"
	m.Tags["disabled"] = "1"
	ok, parts = agent.IsConsulServer(m)
	if !ok {
		t.Fatalf("expected a valid consul server")
	}
	if !parts.Bootstrap {
		t.Fatalf("expected bootstrap")
	}
	if parts.Addr.String() != "127.0.0.1:10000" {
		t.Fatalf("bad addr: %v", parts.Addr)
	}
	if parts.Version != 1 {
		t.Fatalf("bad: %v", parts)
	}
	m.Tags["expect"] = "3"
	delete(m.Tags, "bootstrap")
	delete(m.Tags, "disabled")
	ok, parts = agent.IsConsulServer(m)
	if !ok || parts.Expect != 3 {
		t.Fatalf("bad: %v", parts.Expect)
	}
	if parts.Bootstrap {
		t.Fatalf("unexpected bootstrap")
	}

	delete(m.Tags, "role")
	ok, parts = agent.IsConsulServer(m)
	if ok {
		t.Fatalf("unexpected ok server")
	}
}
