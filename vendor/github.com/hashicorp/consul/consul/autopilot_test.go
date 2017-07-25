package consul

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

func TestAutopilot_CleanupDeadServer(t *testing.T) {
	for i := 1; i <= 3; i++ {
		testCleanupDeadServer(t, i)
	}
}

func testCleanupDeadServer(t *testing.T, raftVersion int) {
	conf := func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = false
		c.BootstrapExpect = 3
		c.RaftConfig.ProtocolVersion = raft.ProtocolVersion(raftVersion)
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
			t.Fatal(err)
		}
	}

	// Bring up a new server
	dir4, s4 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()

	// Kill a non-leader server
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
		t.Fatal(err)
	}

	// Join the new server
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
			t.Fatal(err)
		}
	}
}

func TestAutopilot_CleanupDeadServerPeriodic(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	conf := func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = false
	}
	dir2, s2 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	dir4, s4 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()

	servers := []*Server{s1, s2, s3, s4}

	// Join the servers to s1
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)

	for _, s := range servers[1:] {
		if _, err := s.JoinLAN([]string{addr}); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 4, nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Kill a non-leader server
	s4.Shutdown()

	// Should be removed from the peers automatically
	for _, s := range []*Server{s1, s2, s3} {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAutopilot_CleanupStaleRaftServer(t *testing.T) {
	dir1, s1 := testServerDCBootstrap(t, "dc1", true)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	dir4, s4 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()

	servers := []*Server{s1, s2, s3}

	// Join the servers to s1
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)

	for _, s := range servers[1:] {
		if _, err := s.JoinLAN([]string{addr}); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	for _, s := range servers {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Add s4 to peers directly
	s4addr := fmt.Sprintf("127.0.0.1:%d",
		s4.config.SerfLANConfig.MemberlistConfig.BindPort)
	s1.raft.AddVoter(raft.ServerID(s4.config.NodeID), raft.ServerAddress(s4addr), 0, 0)

	// Verify we have 4 peers
	peers, err := s1.numPeers()
	if err != nil {
		t.Fatal(err)
	}
	if peers != 4 {
		t.Fatalf("bad: %v", peers)
	}

	// Wait for s4 to be removed
	for _, s := range []*Server{s1, s2, s3} {
		if err := testutil.WaitForResult(func() (bool, error) {
			peers, _ := s.numPeers()
			return peers == 3, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAutopilot_PromoteNonVoter(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = true
		c.RaftConfig.ProtocolVersion = 3
		c.AutopilotConfig.ServerStabilizationTime = 200 * time.Millisecond
		c.ServerHealthInterval = 100 * time.Millisecond
		c.AutopilotInterval = 100 * time.Millisecond
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = false
		c.RaftConfig.ProtocolVersion = 3
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Wait for the new server to be added as a non-voter, but make sure
	// it doesn't get promoted to a voter even after ServerStabilizationTime,
	// because that would result in an even-numbered quorum count.
	if err := testutil.WaitForResult(func() (bool, error) {
		future := s1.raft.GetConfiguration()
		if err := future.Error(); err != nil {
			return false, err
		}

		servers := future.Configuration().Servers

		if len(servers) != 2 {
			return false, fmt.Errorf("bad: %v", servers)
		}
		if servers[1].Suffrage != raft.Nonvoter {
			return false, fmt.Errorf("bad: %v", servers)
		}
		health := s1.getServerHealth(string(servers[1].ID))
		if health == nil {
			return false, fmt.Errorf("nil health")
		}
		if !health.Healthy {
			return false, fmt.Errorf("bad: %v", health)
		}
		if time.Now().Sub(health.StableSince) < s1.config.AutopilotConfig.ServerStabilizationTime {
			return false, fmt.Errorf("stable period not elapsed")
		}

		return true, nil
	}); err != nil {
		t.Fatal(err)
	}

	// Now add another server and make sure they both get promoted to voters after stabilization
	dir3, s3 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = false
		c.RaftConfig.ProtocolVersion = 3
	})
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		future := s1.raft.GetConfiguration()
		if err := future.Error(); err != nil {
			return false, err
		}

		servers := future.Configuration().Servers

		if len(servers) != 3 {
			return false, fmt.Errorf("bad: %v", servers)
		}
		if servers[1].Suffrage != raft.Voter {
			return false, fmt.Errorf("bad: %v", servers)
		}
		if servers[2].Suffrage != raft.Voter {
			return false, fmt.Errorf("bad: %v", servers)
		}

		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}
