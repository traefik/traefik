package consul

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/net-rpc-msgpackrpc"
	"github.com/hashicorp/raft"
)

func TestOperator_Autopilot_GetConfiguration(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.AutopilotConfig.CleanupDeadServers = false
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.DCSpecificRequest{
		Datacenter: "dc1",
	}
	var reply structs.AutopilotConfig
	err := msgpackrpc.CallWithCodec(codec, "Operator.AutopilotGetConfiguration", &arg, &reply)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if reply.CleanupDeadServers {
		t.Fatalf("bad: %#v", reply)
	}
}

func TestOperator_Autopilot_GetConfiguration_ACLDeny(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.AutopilotConfig.CleanupDeadServers = false
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Try to get config without permissions
	arg := structs.DCSpecificRequest{
		Datacenter: "dc1",
	}
	var reply structs.AutopilotConfig
	err := msgpackrpc.CallWithCodec(codec, "Operator.AutopilotGetConfiguration", &arg, &reply)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("err: %v", err)
	}

	// Create an ACL with operator read permissions.
	var token string
	{
		var rules = `
                    operator = "read"
                `

		req := structs.ACLRequest{
			Datacenter: "dc1",
			Op:         structs.ACLSet,
			ACL: structs.ACL{
				Name:  "User token",
				Type:  structs.ACLTypeClient,
				Rules: rules,
			},
			WriteRequest: structs.WriteRequest{Token: "root"},
		}
		if err := msgpackrpc.CallWithCodec(codec, "ACL.Apply", &req, &token); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Now we can read and verify the config
	arg.Token = token
	err = msgpackrpc.CallWithCodec(codec, "Operator.AutopilotGetConfiguration", &arg, &reply)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if reply.CleanupDeadServers {
		t.Fatalf("bad: %#v", reply)
	}
}

func TestOperator_Autopilot_SetConfiguration(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.AutopilotConfig.CleanupDeadServers = false
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Change the autopilot config from the default
	arg := structs.AutopilotSetConfigRequest{
		Datacenter: "dc1",
		Config: structs.AutopilotConfig{
			CleanupDeadServers: true,
		},
	}
	var reply *bool
	err := msgpackrpc.CallWithCodec(codec, "Operator.AutopilotSetConfiguration", &arg, &reply)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make sure it's changed
	state := s1.fsm.State()
	_, config, err := state.AutopilotConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !config.CleanupDeadServers {
		t.Fatalf("bad: %#v", config)
	}
}

func TestOperator_Autopilot_SetConfiguration_ACLDeny(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.AutopilotConfig.CleanupDeadServers = false
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Try to set config without permissions
	arg := structs.AutopilotSetConfigRequest{
		Datacenter: "dc1",
		Config: structs.AutopilotConfig{
			CleanupDeadServers: true,
		},
	}
	var reply *bool
	err := msgpackrpc.CallWithCodec(codec, "Operator.AutopilotSetConfiguration", &arg, &reply)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("err: %v", err)
	}

	// Create an ACL with operator write permissions.
	var token string
	{
		var rules = `
                    operator = "write"
                `

		req := structs.ACLRequest{
			Datacenter: "dc1",
			Op:         structs.ACLSet,
			ACL: structs.ACL{
				Name:  "User token",
				Type:  structs.ACLTypeClient,
				Rules: rules,
			},
			WriteRequest: structs.WriteRequest{Token: "root"},
		}
		if err := msgpackrpc.CallWithCodec(codec, "ACL.Apply", &req, &token); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Now we can update the config
	arg.Token = token
	err = msgpackrpc.CallWithCodec(codec, "Operator.AutopilotSetConfiguration", &arg, &reply)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make sure it's changed
	state := s1.fsm.State()
	_, config, err := state.AutopilotConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !config.CleanupDeadServers {
		t.Fatalf("bad: %#v", config)
	}
}

func TestOperator_ServerHealth(t *testing.T) {
	conf := func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = false
		c.BootstrapExpect = 3
		c.RaftConfig.ProtocolVersion = 3
		c.ServerHealthInterval = 100 * time.Millisecond
		c.AutopilotInterval = 100 * time.Millisecond
	}
	dir1, s1 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	dir2, s2 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	dir3, s3 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()
	if _, err := s3.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	if err := testutil.WaitForResult(func() (bool, error) {
		arg := structs.DCSpecificRequest{
			Datacenter: "dc1",
		}
		var reply structs.OperatorHealthReply
		err := msgpackrpc.CallWithCodec(codec, "Operator.ServerHealth", &arg, &reply)
		if err != nil {
			return false, fmt.Errorf("err: %v", err)
		}
		if !reply.Healthy {
			return false, fmt.Errorf("bad: %v", reply)
		}
		if reply.FailureTolerance != 1 {
			return false, fmt.Errorf("bad: %v", reply)
		}
		if len(reply.Servers) != 3 {
			return false, fmt.Errorf("bad: %v", reply)
		}
		// Leader should have LastContact == 0, others should be positive
		for _, s := range reply.Servers {
			isLeader := s1.raft.Leader() == raft.ServerAddress(s.Address)
			if isLeader && s.LastContact != 0 {
				return false, fmt.Errorf("bad: %v", reply)
			}
			if !isLeader && s.LastContact <= 0 {
				return false, fmt.Errorf("bad: %v", reply)
			}
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestOperator_ServerHealth_UnsupportedRaftVersion(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc1"
		c.Bootstrap = true
		c.RaftConfig.ProtocolVersion = 2
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	arg := structs.DCSpecificRequest{
		Datacenter: "dc1",
	}
	var reply structs.OperatorHealthReply
	err := msgpackrpc.CallWithCodec(codec, "Operator.ServerHealth", &arg, &reply)
	if err == nil || !strings.Contains(err.Error(), "raft_protocol set to 3 or higher") {
		t.Fatalf("bad: %v", err)
	}
}
