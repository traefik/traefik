package consul

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/consul/types"
)

func TestStatsFetcher(t *testing.T) {
	dir1, s1 := testServerDCExpect(t, "dc1", 3)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerDCExpect(t, "dc1", 3)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCExpect(t, "dc1", 3)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")

	members := s1.serfLAN.Members()
	if len(members) != 3 {
		t.Fatalf("bad len: %d", len(members))
	}

	var servers []*agent.Server
	for _, member := range members {
		ok, server := agent.IsConsulServer(member)
		if !ok {
			t.Fatalf("bad: %#v", member)
		}
		servers = append(servers, server)
	}

	// Do a normal fetch and make sure we get three responses.
	func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		stats := s1.statsFetcher.Fetch(ctx, servers)
		if len(stats) != 3 {
			t.Fatalf("bad: %#v", stats)
		}
		for id, stat := range stats {
			switch types.NodeID(id) {
			case s1.config.NodeID, s2.config.NodeID, s3.config.NodeID:
				// OK
			default:
				t.Fatalf("bad: %s", id)
			}

			if stat == nil || stat.LastTerm == 0 {
				t.Fatalf("bad: %#v", stat)
			}
		}
	}()

	// Fake an in-flight request to server 3 and make sure we don't fetch
	// from it.
	func() {
		s1.statsFetcher.inflight[string(s3.config.NodeID)] = struct{}{}
		defer delete(s1.statsFetcher.inflight, string(s3.config.NodeID))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		stats := s1.statsFetcher.Fetch(ctx, servers)
		if len(stats) != 2 {
			t.Fatalf("bad: %#v", stats)
		}
		for id, stat := range stats {
			switch types.NodeID(id) {
			case s1.config.NodeID, s2.config.NodeID:
				// OK
			case s3.config.NodeID:
				t.Fatalf("bad")
			default:
				t.Fatalf("bad: %s", id)
			}

			if stat == nil || stat.LastTerm == 0 {
				t.Fatalf("bad: %#v", stat)
			}
		}
	}()

	// Do a fetch with a canceled context and make sure we bail right away.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()
	stats := s1.statsFetcher.Fetch(ctx, servers)
	if len(stats) != 0 {
		t.Fatalf("bad: %#v", stats)
	}
}
