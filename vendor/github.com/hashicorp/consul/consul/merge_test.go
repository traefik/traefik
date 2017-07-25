package consul

import (
	"strings"
	"testing"

	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/serf/serf"
)

func makeNode(dc, name, id string, server bool) *serf.Member {
	var role string
	if server {
		role = "consul"
	} else {
		role = "node"
	}

	return &serf.Member{
		Name: name,
		Tags: map[string]string{
			"role":    role,
			"dc":      dc,
			"id":      id,
			"port":    "8300",
			"build":   "0.7.5",
			"vsn":     "2",
			"vsn_max": "3",
			"vsn_min": "2",
		},
	}
}

func TestMerge_LAN(t *testing.T) {
	cases := []struct {
		members []*serf.Member
		expect  string
	}{
		// Client in the wrong datacenter.
		{
			members: []*serf.Member{
				makeNode("dc2",
					"node1",
					"96430788-246f-4379-94ce-257f7429e340",
					false),
			},
			expect: "wrong datacenter",
		},
		// Server in the wrong datacenter.
		{
			members: []*serf.Member{
				makeNode("dc2",
					"node1",
					"96430788-246f-4379-94ce-257f7429e340",
					true),
			},
			expect: "wrong datacenter",
		},
		// Node ID conflict with delegate's ID.
		{
			members: []*serf.Member{
				makeNode("dc1",
					"node1",
					"ee954a2f-80de-4b34-8780-97b942a50a99",
					true),
			},
			expect: "with this agent's ID",
		},
		// Cluster with existing conflicting node IDs.
		{
			members: []*serf.Member{
				makeNode("dc1",
					"node1",
					"6185913b-98d7-4441-bd8f-f7f7d854a4af",
					true),
				makeNode("dc1",
					"node2",
					"6185913b-98d7-4441-bd8f-f7f7d854a4af",
					true),
			},
			expect: "with member",
		},
		// Good cluster.
		{
			members: []*serf.Member{
				makeNode("dc1",
					"node1",
					"6185913b-98d7-4441-bd8f-f7f7d854a4af",
					true),
				makeNode("dc1",
					"node2",
					"cda916bc-a357-4a19-b886-59419fcee50c",
					true),
			},
			expect: "",
		},
	}

	delegate := &lanMergeDelegate{
		dc:       "dc1",
		nodeID:   types.NodeID("ee954a2f-80de-4b34-8780-97b942a50a99"),
		nodeName: "node0",
	}
	for i, c := range cases {
		if err := delegate.NotifyMerge(c.members); c.expect == "" {
			if err != nil {
				t.Fatalf("case %d: err: %v", i+1, err)
			}
		} else {
			if err == nil || !strings.Contains(err.Error(), c.expect) {
				t.Fatalf("case %d: err: %v", i+1, err)
			}
		}
	}
}

func TestMerge_WAN(t *testing.T) {
	cases := []struct {
		members []*serf.Member
		expect  string
	}{
		// Not a server
		{
			members: []*serf.Member{
				makeNode("dc2",
					"node1",
					"96430788-246f-4379-94ce-257f7429e340",
					false),
			},
			expect: "not a server",
		},
		// Good cluster.
		{
			members: []*serf.Member{
				makeNode("dc2",
					"node1",
					"6185913b-98d7-4441-bd8f-f7f7d854a4af",
					true),
				makeNode("dc3",
					"node2",
					"cda916bc-a357-4a19-b886-59419fcee50c",
					true),
			},
			expect: "",
		},
	}

	delegate := &wanMergeDelegate{}
	for i, c := range cases {
		if err := delegate.NotifyMerge(c.members); c.expect == "" {
			if err != nil {
				t.Fatalf("case %d: err: %v", i+1, err)
			}
		} else {
			if err == nil || !strings.Contains(err.Error(), c.expect) {
				t.Fatalf("case %d: err: %v", i+1, err)
			}
		}
	}
}
