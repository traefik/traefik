package consul

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/net-rpc-msgpackrpc"
	"github.com/hashicorp/serf/serf"
)

func TestLeader_RegisterMember(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, c1 := testClient(t)
	defer os.RemoveAll(dir2)
	defer c1.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := c1.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Client should be registered
	state := s1.fsm.State()
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err := state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return node != nil, nil
	}); err != nil {
		t.Fatal("client not registered")
	}

	// Should have a check
	_, checks, err := state.NodeChecks(nil, c1.config.NodeName)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(checks) != 1 {
		t.Fatalf("client missing check")
	}
	if checks[0].CheckID != SerfCheckID {
		t.Fatalf("bad check: %v", checks[0])
	}
	if checks[0].Name != SerfCheckName {
		t.Fatalf("bad check: %v", checks[0])
	}
	if checks[0].Status != structs.HealthPassing {
		t.Fatalf("bad check: %v", checks[0])
	}

	// Server should be registered
	_, node, err := state.GetNode(s1.config.NodeName)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if node == nil {
		t.Fatalf("server not registered")
	}

	// Service should be registered
	_, services, err := state.NodeServices(nil, s1.config.NodeName)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, ok := services.Services["consul"]; !ok {
		t.Fatalf("consul service not registered: %v", services)
	}
}

func TestLeader_FailedMember(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, c1 := testClient(t)
	defer os.RemoveAll(dir2)
	defer c1.Shutdown()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := c1.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Fail the member
	c1.Shutdown()

	// Should be registered
	state := s1.fsm.State()
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err := state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return node != nil, nil
	}); err != nil {
		t.Fatal("client not registered")
	}

	// Should have a check
	_, checks, err := state.NodeChecks(nil, c1.config.NodeName)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(checks) != 1 {
		t.Fatalf("client missing check")
	}
	if checks[0].CheckID != SerfCheckID {
		t.Fatalf("bad check: %v", checks[0])
	}
	if checks[0].Name != SerfCheckName {
		t.Fatalf("bad check: %v", checks[0])
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		_, checks, err = state.NodeChecks(nil, c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return checks[0].Status == structs.HealthCritical, errors.New(checks[0].Status)
	}); err != nil {
		t.Fatalf("check status is %v, should be critical", err)
	}
}

func TestLeader_LeftMember(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, c1 := testClient(t)
	defer os.RemoveAll(dir2)
	defer c1.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := c1.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	state := s1.fsm.State()

	// Should be registered
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err := state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return node != nil, nil
	}); err != nil {
		t.Fatal("client should be registered")
	}

	// Node should leave
	c1.Leave()
	c1.Shutdown()

	// Should be deregistered
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err := state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return node == nil, nil
	}); err != nil {
		t.Fatal("client should not be registered")
	}
}

func TestLeader_ReapMember(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, c1 := testClient(t)
	defer os.RemoveAll(dir2)
	defer c1.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := c1.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	state := s1.fsm.State()

	// Should be registered
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err := state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return node != nil, nil
	}); err != nil {
		t.Fatal("client should be registered")
	}

	// Simulate a node reaping
	mems := s1.LANMembers()
	var c1mem serf.Member
	for _, m := range mems {
		if m.Name == c1.config.NodeName {
			c1mem = m
			c1mem.Status = StatusReap
			break
		}
	}
	s1.reconcileCh <- c1mem

	// Should be deregistered; we have to poll quickly here because
	// anti-entropy will put it back.
	reaped := false
	for start := time.Now(); time.Since(start) < 5*time.Second; {
		_, node, err := state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if node == nil {
			reaped = true
			break
		}
	}
	if !reaped {
		t.Fatalf("client should not be registered")
	}
}

func TestLeader_Reconcile_ReapMember(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Register a non-existing member
	dead := structs.RegisterRequest{
		Datacenter: s1.config.Datacenter,
		Node:       "no-longer-around",
		Address:    "127.1.1.1",
		Check: &structs.HealthCheck{
			Node:    "no-longer-around",
			CheckID: SerfCheckID,
			Name:    SerfCheckName,
			Status:  structs.HealthCritical,
		},
		WriteRequest: structs.WriteRequest{
			Token: "root",
		},
	}
	var out struct{}
	if err := s1.RPC("Catalog.Register", &dead, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Force a reconciliation
	if err := s1.reconcile(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Node should be gone
	state := s1.fsm.State()
	_, node, err := state.GetNode("no-longer-around")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if node != nil {
		t.Fatalf("client registered")
	}
}

func TestLeader_Reconcile(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, c1 := testClient(t)
	defer os.RemoveAll(dir2)
	defer c1.Shutdown()

	// Join before we have a leader, this should cause a reconcile!
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := c1.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Should not be registered
	state := s1.fsm.State()
	_, node, err := state.GetNode(c1.config.NodeName)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if node != nil {
		t.Fatalf("client registered")
	}

	// Should be registered
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err = state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return node != nil, nil
	}); err != nil {
		t.Fatal("client should be registered")
	}
}

func TestLeader_Reconcile_Races(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	dir2, c1 := testClient(t)
	defer os.RemoveAll(dir2)
	defer c1.Shutdown()

	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := c1.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait for the server to reconcile the client and register it.
	state := s1.fsm.State()
	var nodeAddr string
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err := state.GetNode(c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if node != nil {
			nodeAddr = node.Address
			return true, nil
		} else {
			return false, nil
		}
	}); err != nil {
		t.Fatalf("client should be registered: %v", err)
	}

	// Add in some metadata via the catalog (as if the agent synced it
	// there). We also set the serfHealth check to failing so the reconile
	// will attempt to flip it back
	req := structs.RegisterRequest{
		Datacenter: s1.config.Datacenter,
		Node:       c1.config.NodeName,
		ID:         c1.config.NodeID,
		Address:    nodeAddr,
		NodeMeta:   map[string]string{"hello": "world"},
		Check: &structs.HealthCheck{
			Node:    c1.config.NodeName,
			CheckID: SerfCheckID,
			Name:    SerfCheckName,
			Status:  structs.HealthCritical,
			Output:  "",
		},
	}
	var out struct{}
	if err := s1.RPC("Catalog.Register", &req, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Force a reconcile and make sure the metadata stuck around.
	if err := s1.reconcile(); err != nil {
		t.Fatalf("err: %v", err)
	}
	_, node, err := state.GetNode(c1.config.NodeName)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if node == nil {
		t.Fatalf("bad")
	}
	if hello, ok := node.Meta["hello"]; !ok || hello != "world" {
		t.Fatalf("bad")
	}

	// Fail the member and wait for the health to go critical.
	c1.Shutdown()
	if err := testutil.WaitForResult(func() (bool, error) {
		_, checks, err := state.NodeChecks(nil, c1.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return checks[0].Status == structs.HealthCritical, errors.New(checks[0].Status)
	}); err != nil {
		t.Fatalf("check status should be critical: %v", err)
	}

	// Make sure the metadata didn't get clobbered.
	_, node, err = state.GetNode(c1.config.NodeName)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if node == nil {
		t.Fatalf("bad")
	}
	if hello, ok := node.Meta["hello"]; !ok || hello != "world" {
		t.Fatalf("bad")
	}
}

func TestLeader_LeftServer(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()
	servers := []*Server{s1, s2, s3}

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal("should have 3 peers")
		}
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		// Kill any server
		servers[0].Shutdown()

		// Force remove the non-leader (transition to left state)
		if err := servers[1].RemoveFailedNode(servers[0].config.NodeName); err != nil {
			t.Fatalf("err: %v", err)
		}

		for _, s := range servers[1:] {
			peers, _ := s.numPeers()
			return peers == 2, errors.New(fmt.Sprintf("%d", peers))
		}

		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLeader_LeftLeader(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()
	servers := []*Server{s1, s2, s3}

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal("should have 3 peers")
		}
	}

	// Kill the leader!
	var leader *Server
	for _, s := range servers {
		if s.IsLeader() {
			leader = s
			break
		}
	}
	if leader == nil {
		t.Fatalf("Should have a leader")
	}
	leader.Leave()
	leader.Shutdown()
	time.Sleep(100 * time.Millisecond)

	var remain *Server
	for _, s := range servers {
		if s == leader {
			continue
		}
		remain = s
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 2, errors.New(fmt.Sprintf("%d", peers))
		}); err != nil {
			t.Fatal("should have 2 peers")
		}
	}

	// Verify the old leader is deregistered
	state := remain.fsm.State()
	if err := testutil.WaitForResult(func() (bool, error) {
		_, node, err := state.GetNode(leader.config.NodeName)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		return node == nil, nil
	}); err != nil {
		t.Fatal("should be deregistered")
	}
}

func TestLeader_MultiBootstrap(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServer(t)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	servers := []*Server{s1, s2}

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers := s.serfLAN.Members()
			return len(peers) == 2, nil
		}); err != nil {
			t.Fatal("should have 2 peerss")
		}
	}

	// Ensure we don't have multiple raft peers
	for _, s := range servers {
		peers, _ := s.numPeers()
		if peers != 1 {
			t.Fatalf("should only have 1 raft peer!")
		}
	}
}

func TestLeader_TombstoneGC_Reset(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()
	servers := []*Server{s1, s2, s3}

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal("should have 3 peers")
		}
	}

	var leader *Server
	for _, s := range servers {
		if s.IsLeader() {
			leader = s
			break
		}
	}
	if leader == nil {
		t.Fatalf("Should have a leader")
	}

	// Check that the leader has a pending GC expiration
	if !leader.tombstoneGC.PendingExpiration() {
		t.Fatalf("should have pending expiration")
	}

	// Kill the leader
	leader.Shutdown()
	time.Sleep(100 * time.Millisecond)

	// Wait for a new leader
	leader = nil
	if err := testutil.WaitForResult(func() (bool, error) {
		for _, s := range servers {
			if s.IsLeader() {
				leader = s
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		t.Fatal("should have leader")
	}

	// Check that the new leader has a pending GC expiration
	if err := testutil.WaitForResult(func() (bool, error) {
		return leader.tombstoneGC.PendingExpiration(), nil
	}); err != nil {
		t.Fatal("should have pending expiration")
	}
}

func TestLeader_ReapTombstones(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.TombstoneTTL = 50 * time.Millisecond
		c.TombstoneTTLGranularity = 10 * time.Millisecond
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create a KV entry
	arg := structs.KVSRequest{
		Datacenter: "dc1",
		Op:         structs.KVSSet,
		DirEnt: structs.DirEntry{
			Key:   "test",
			Value: []byte("test"),
		},
		WriteRequest: structs.WriteRequest{
			Token: "root",
		},
	}
	var out bool
	if err := msgpackrpc.CallWithCodec(codec, "KVS.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Delete the KV entry (tombstoned).
	arg.Op = structs.KVSDelete
	if err := msgpackrpc.CallWithCodec(codec, "KVS.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make sure there's a tombstone.
	state := s1.fsm.State()
	func() {
		snap := state.Snapshot()
		defer snap.Close()
		stones, err := snap.Tombstones()
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if stones.Next() == nil {
			t.Fatalf("missing tombstones")
		}
		if stones.Next() != nil {
			t.Fatalf("unexpected extra tombstones")
		}
	}()

	// Check that the new leader has a pending GC expiration by
	// watching for the tombstone to get removed.
	if err := testutil.WaitForResult(func() (bool, error) {
		snap := state.Snapshot()
		defer snap.Close()
		stones, err := snap.Tombstones()
		if err != nil {
			return false, err
		}
		return stones.Next() == nil, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLeader_RollRaftServer(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = true
		c.Datacenter = "dc1"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.Datacenter = "dc1"
		c.RaftConfig.ProtocolVersion = 1
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	servers := []*Server{s1, s2, s3}

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal("should have 3 peers")
		}
	}

	// Kill the v1 server
	s2.Shutdown()

	for _, s := range []*Server{s1, s3} {
		if err := testutil.WaitForResult(func() (bool, error) {
			minVer, err := ServerMinRaftProtocol(s.LANMembers())
			return minVer == 2, err
		}); err != nil {
			t.Fatalf("minimum protocol version among servers should be 2")
		}
	}

	// Replace the dead server with one running raft protocol v3
	dir4, s4 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.Datacenter = "dc1"
		c.RaftConfig.ProtocolVersion = 3
	})
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()
	if _, err := s4.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	servers[1] = s4

	// Make sure the dead server is removed and we're back to 3 total peers
	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			addrs := 0
			ids := 0
			future := s.raft.GetConfiguration()
			if err := future.Error(); err != nil {
				return false, err
			}
			for _, server := range future.Configuration().Servers {
				if string(server.ID) == string(server.Address) {
					addrs++
				} else {
					ids++
				}
			}
			return addrs == 2 && ids == 1, nil
		}); err != nil {
			t.Fatalf("should see 2 legacy IDs and 1 GUID")
		}
	}
}

func TestLeader_ChangeServerID(t *testing.T) {
	conf := func(c *Config) {
		c.Bootstrap = false
		c.BootstrapExpect = 3
		c.Datacenter = "dc1"
		c.RaftConfig.ProtocolVersion = 3
	}
	dir1, s1 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	servers := []*Server{s1, s2, s3}

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal("should have 3 peers")
		}
	}

	// Shut down a server, freeing up its address/port
	s3.Shutdown()

	if err := testutil.WaitForResult(func() (bool, error) {
		alive := 0
		for _, m := range s1.LANMembers() {
			if m.Status == serf.StatusAlive {
				alive++
			}
		}
		return alive == 2, nil
	}); err != nil {
		t.Fatal("should have 2 alive members")
	}

	// Bring up a new server with s3's address that will get a different ID
	dir4, s4 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.BootstrapExpect = 3
		c.Datacenter = "dc1"
		c.RaftConfig.ProtocolVersion = 3
		c.SerfLANConfig.MemberlistConfig = s3.config.SerfLANConfig.MemberlistConfig
		c.RPCAddr = s3.config.RPCAddr
		c.RPCAdvertise = s3.config.RPCAdvertise
	})
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()
	if _, err := s4.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	servers[2] = s4

	// Make sure the dead server is removed and we're back to 3 total peers
	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal("should have 3 peers")
		}
	}
}
