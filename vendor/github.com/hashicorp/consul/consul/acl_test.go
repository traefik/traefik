package consul

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/testutil"
)

var testACLPolicy = `
key "" {
	policy = "deny"
}
key "foo/" {
	policy = "write"
}
`

func TestACL_Disabled(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	acl, err := s1.resolveToken("does not exist")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl != nil {
		t.Fatalf("got acl")
	}
}

func TestACL_ResolveRootACL(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	acl, err := s1.resolveToken("allow")
	if err == nil || err.Error() != rootDenied {
		t.Fatalf("err: %v", err)
	}
	if acl != nil {
		t.Fatalf("bad: %v", acl)
	}

	acl, err = s1.resolveToken("deny")
	if err == nil || err.Error() != rootDenied {
		t.Fatalf("err: %v", err)
	}
	if acl != nil {
		t.Fatalf("bad: %v", acl)
	}
}

func TestACL_Authority_NotFound(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	acl, err := s1.resolveToken("does not exist")
	if err == nil || err.Error() != aclNotFound {
		t.Fatalf("err: %v", err)
	}
	if acl != nil {
		t.Fatalf("got acl")
	}
}

func TestACL_Authority_Found(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLMasterToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create a new token
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name:  "User token",
			Type:  structs.ACLTypeClient,
			Rules: testACLPolicy,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var id string
	if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Resolve the token
	acl, err := s1.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy
	if acl.KeyRead("bar") {
		t.Fatalf("unexpected read")
	}
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}
}

func TestACL_Authority_Anonymous_Found(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Resolve the token
	acl, err := s1.resolveToken("")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy, should allow all
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}
}

func TestACL_Authority_Master_Found(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLMasterToken = "foobar"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Resolve the token
	acl, err := s1.resolveToken("foobar")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy, should allow all
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}
}

func TestACL_Authority_Management(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLMasterToken = "foobar"
		c.ACLDefaultPolicy = "deny"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Resolve the token
	acl, err := s1.resolveToken("foobar")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy, should allow all
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}
}

func TestACL_NonAuthority_NotFound(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.Bootstrap = false     // Disable bootstrap
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		p1, _ := s1.numPeers()
		return p1 == 2, errors.New(fmt.Sprintf("%d", p1))
	}); err != nil {
		t.Fatal(err)
	}

	client := rpcClient(t, s1)
	defer client.Close()
	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// find the non-authoritative server
	var nonAuth *Server
	if !s1.IsLeader() {
		nonAuth = s1
	} else {
		nonAuth = s2
	}

	acl, err := nonAuth.resolveToken("does not exist")
	if err == nil || err.Error() != aclNotFound {
		t.Fatalf("err: %v", err)
	}
	if acl != nil {
		t.Fatalf("got acl")
	}
}

func TestACL_NonAuthority_Found(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.Bootstrap = false     // Disable bootstrap
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		p1, _ := s1.numPeers()
		return p1 == 2, errors.New(fmt.Sprintf("%d", p1))
	}); err != nil {
		t.Fatal(err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create a new token
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name:  "User token",
			Type:  structs.ACLTypeClient,
			Rules: testACLPolicy,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var id string
	if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// find the non-authoritative server
	var nonAuth *Server
	if !s1.IsLeader() {
		nonAuth = s1
	} else {
		nonAuth = s2
	}

	// Token should resolve
	acl, err := nonAuth.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy
	if acl.KeyRead("bar") {
		t.Fatalf("unexpected read")
	}
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}
}

func TestACL_NonAuthority_Management(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLMasterToken = "foobar"
		c.ACLDefaultPolicy = "deny"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLDefaultPolicy = "deny"
		c.Bootstrap = false // Disable bootstrap
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		p1, _ := s1.numPeers()
		return p1 == 2, errors.New(fmt.Sprintf("%d", p1))
	}); err != nil {
		t.Fatal(err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// find the non-authoritative server
	var nonAuth *Server
	if !s1.IsLeader() {
		nonAuth = s1
	} else {
		nonAuth = s2
	}

	// Resolve the token
	acl, err := nonAuth.resolveToken("foobar")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy, should allow all
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}
}

func TestACL_DownPolicy_Deny(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLDownPolicy = "deny"
		c.ACLMasterToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLDownPolicy = "deny"
		c.Bootstrap = false // Disable bootstrap
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		p1, _ := s1.numPeers()
		return p1 == 2, errors.New(fmt.Sprintf("%d", p1))
	}); err != nil {
		t.Fatal(err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create a new token
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name:  "User token",
			Type:  structs.ACLTypeClient,
			Rules: testACLPolicy,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var id string
	if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// find the non-authoritative server
	var nonAuth *Server
	var auth *Server
	if !s1.IsLeader() {
		nonAuth = s1
		auth = s2
	} else {
		nonAuth = s2
		auth = s1
	}

	// Kill the authoritative server
	auth.Shutdown()

	// Token should resolve into a DenyAll
	aclR, err := nonAuth.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if aclR != acl.DenyAll() {
		t.Fatalf("bad acl: %#v", aclR)
	}
}

func TestACL_DownPolicy_Allow(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLDownPolicy = "allow"
		c.ACLMasterToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLDownPolicy = "allow"
		c.Bootstrap = false // Disable bootstrap
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		p1, _ := s1.numPeers()
		return p1 == 2, errors.New(fmt.Sprintf("%d", p1))
	}); err != nil {
		t.Fatal(err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create a new token
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name:  "User token",
			Type:  structs.ACLTypeClient,
			Rules: testACLPolicy,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var id string
	if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// find the non-authoritative server
	var nonAuth *Server
	var auth *Server
	if !s1.IsLeader() {
		nonAuth = s1
		auth = s2
	} else {
		nonAuth = s2
		auth = s1
	}

	// Kill the authoritative server
	auth.Shutdown()

	// Token should resolve into a AllowAll
	aclR, err := nonAuth.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if aclR != acl.AllowAll() {
		t.Fatalf("bad acl: %#v", aclR)
	}
}

func TestACL_DownPolicy_ExtendCache(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLTTL = 0
		c.ACLDownPolicy = "extend-cache"
		c.ACLMasterToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1" // Enable ACLs!
		c.ACLTTL = 0
		c.ACLDownPolicy = "extend-cache"
		c.Bootstrap = false // Disable bootstrap
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfLANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinLAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		p1, _ := s1.numPeers()
		return p1 == 2, errors.New(fmt.Sprintf("%d", p1))
	}); err != nil {
		t.Fatal(err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create a new token
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name:  "User token",
			Type:  structs.ACLTypeClient,
			Rules: testACLPolicy,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var id string
	if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// find the non-authoritative server
	var nonAuth *Server
	var auth *Server
	if !s1.IsLeader() {
		nonAuth = s1
		auth = s2
	} else {
		nonAuth = s2
		auth = s1
	}

	// Warm the caches
	aclR, err := nonAuth.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if aclR == nil {
		t.Fatalf("bad acl: %#v", aclR)
	}

	// Kill the authoritative server
	auth.Shutdown()

	// Token should resolve into cached copy
	aclR2, err := nonAuth.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if aclR2 != aclR {
		t.Fatalf("bad acl: %#v", aclR)
	}
}

func TestACL_Replication(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc2"
		c.ACLDatacenter = "dc1"
		c.ACLDefaultPolicy = "deny"
		c.ACLDownPolicy = "extend-cache"
		c.ACLReplicationToken = "root"
		c.ACLReplicationInterval = 0
		c.ACLReplicationApplyLimit = 1000000
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc3"
		c.ACLDatacenter = "dc1"
		c.ACLDownPolicy = "deny"
		c.ACLReplicationToken = "root"
		c.ACLReplicationInterval = 0
		c.ACLReplicationApplyLimit = 1000000
	})
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	// Try to join.
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfWANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinWAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := s3.JoinWAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")
	testutil.WaitForLeader(t, s1.RPC, "dc2")
	testutil.WaitForLeader(t, s1.RPC, "dc3")

	// Create a new token.
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name:  "User token",
			Type:  structs.ACLTypeClient,
			Rules: testACLPolicy,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var id string
	if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait for replication to occur.
	if err := testutil.WaitForResult(func() (bool, error) {
		_, acl, err := s2.fsm.State().ACLGet(nil, id)
		if err != nil {
			return false, err
		}
		if acl == nil {
			return false, nil
		}
		_, acl, err = s3.fsm.State().ACLGet(nil, id)
		if err != nil {
			return false, err
		}
		if acl == nil {
			return false, nil
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}

	// Kill the ACL datacenter.
	s1.Shutdown()

	// Token should resolve on s2, which has replication + extend-cache.
	acl, err := s2.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy
	if acl.KeyRead("bar") {
		t.Fatalf("unexpected read")
	}
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}

	// Although s3 has replication, and we verified that the ACL is there,
	// it can not be used because of the down policy.
	acl, err = s3.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy.
	if acl.KeyRead("bar") {
		t.Fatalf("unexpected read")
	}
	if acl.KeyRead("foo/test") {
		t.Fatalf("unexpected read")
	}
}

func TestACL_MultiDC_Found(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	client := rpcClient(t, s1)
	defer client.Close()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc2"
		c.ACLDatacenter = "dc1" // Enable ACLs!
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfWANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinWAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}

	testutil.WaitForLeader(t, s1.RPC, "dc1")
	testutil.WaitForLeader(t, s1.RPC, "dc2")

	// Create a new token
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name:  "User token",
			Type:  structs.ACLTypeClient,
			Rules: testACLPolicy,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var id string
	if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Token should resolve
	acl, err := s2.resolveToken(id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if acl == nil {
		t.Fatalf("missing acl")
	}

	// Check the policy
	if acl.KeyRead("bar") {
		t.Fatalf("unexpected read")
	}
	if !acl.KeyRead("foo/test") {
		t.Fatalf("unexpected failed read")
	}
}

func TestACL_filterHealthChecks(t *testing.T) {
	// Create some health checks.
	fill := func() structs.HealthChecks {
		return structs.HealthChecks{
			&structs.HealthCheck{
				Node:        "node1",
				CheckID:     "check1",
				ServiceName: "foo",
			},
		}
	}

	// Try permissive filtering.
	{
		hc := fill()
		filt := newAclFilter(acl.AllowAll(), nil, false)
		filt.filterHealthChecks(&hc)
		if len(hc) != 1 {
			t.Fatalf("bad: %#v", hc)
		}
	}

	// Try restrictive filtering.
	{
		hc := fill()
		filt := newAclFilter(acl.DenyAll(), nil, false)
		filt.filterHealthChecks(&hc)
		if len(hc) != 0 {
			t.Fatalf("bad: %#v", hc)
		}
	}

	// Allowed to see the service but not the node.
	policy, err := acl.Parse(`
service "foo" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err := acl.New(acl.DenyAll(), policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// This will work because version 8 ACLs aren't being enforced.
	{
		hc := fill()
		filt := newAclFilter(perms, nil, false)
		filt.filterHealthChecks(&hc)
		if len(hc) != 1 {
			t.Fatalf("bad: %#v", hc)
		}
	}

	// But with version 8 the node will block it.
	{
		hc := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterHealthChecks(&hc)
		if len(hc) != 0 {
			t.Fatalf("bad: %#v", hc)
		}
	}

	// Chain on access to the node.
	policy, err = acl.Parse(`
node "node1" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now it should go through.
	{
		hc := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterHealthChecks(&hc)
		if len(hc) != 1 {
			t.Fatalf("bad: %#v", hc)
		}
	}
}

func TestACL_filterServices(t *testing.T) {
	// Create some services
	services := structs.Services{
		"service1": []string{},
		"service2": []string{},
		"consul":   []string{},
	}

	// Try permissive filtering.
	filt := newAclFilter(acl.AllowAll(), nil, false)
	filt.filterServices(services)
	if len(services) != 3 {
		t.Fatalf("bad: %#v", services)
	}

	// Try restrictive filtering.
	filt = newAclFilter(acl.DenyAll(), nil, false)
	filt.filterServices(services)
	if len(services) != 1 {
		t.Fatalf("bad: %#v", services)
	}
	if _, ok := services["consul"]; !ok {
		t.Fatalf("bad: %#v", services)
	}

	// Try restrictive filtering with version 8 enforcement.
	filt = newAclFilter(acl.DenyAll(), nil, true)
	filt.filterServices(services)
	if len(services) != 0 {
		t.Fatalf("bad: %#v", services)
	}
}

func TestACL_filterServiceNodes(t *testing.T) {
	// Create some service nodes.
	fill := func() structs.ServiceNodes {
		return structs.ServiceNodes{
			&structs.ServiceNode{
				Node:        "node1",
				ServiceName: "foo",
			},
		}
	}

	// Try permissive filtering.
	{
		nodes := fill()
		filt := newAclFilter(acl.AllowAll(), nil, false)
		filt.filterServiceNodes(&nodes)
		if len(nodes) != 1 {
			t.Fatalf("bad: %#v", nodes)
		}
	}

	// Try restrictive filtering.
	{
		nodes := fill()
		filt := newAclFilter(acl.DenyAll(), nil, false)
		filt.filterServiceNodes(&nodes)
		if len(nodes) != 0 {
			t.Fatalf("bad: %#v", nodes)
		}
	}

	// Allowed to see the service but not the node.
	policy, err := acl.Parse(`
service "foo" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err := acl.New(acl.DenyAll(), policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// This will work because version 8 ACLs aren't being enforced.
	{
		nodes := fill()
		filt := newAclFilter(perms, nil, false)
		filt.filterServiceNodes(&nodes)
		if len(nodes) != 1 {
			t.Fatalf("bad: %#v", nodes)
		}
	}

	// But with version 8 the node will block it.
	{
		nodes := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterServiceNodes(&nodes)
		if len(nodes) != 0 {
			t.Fatalf("bad: %#v", nodes)
		}
	}

	// Chain on access to the node.
	policy, err = acl.Parse(`
node "node1" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now it should go through.
	{
		nodes := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterServiceNodes(&nodes)
		if len(nodes) != 1 {
			t.Fatalf("bad: %#v", nodes)
		}
	}
}

func TestACL_filterNodeServices(t *testing.T) {
	// Create some node services.
	fill := func() *structs.NodeServices {
		return &structs.NodeServices{
			Node: &structs.Node{
				Node: "node1",
			},
			Services: map[string]*structs.NodeService{
				"foo": &structs.NodeService{
					ID:      "foo",
					Service: "foo",
				},
			},
		}
	}

	// Try nil, which is a possible input.
	{
		var services *structs.NodeServices
		filt := newAclFilter(acl.AllowAll(), nil, false)
		filt.filterNodeServices(&services)
		if services != nil {
			t.Fatalf("bad: %#v", services)
		}
	}

	// Try permissive filtering.
	{
		services := fill()
		filt := newAclFilter(acl.AllowAll(), nil, false)
		filt.filterNodeServices(&services)
		if len(services.Services) != 1 {
			t.Fatalf("bad: %#v", services.Services)
		}
	}

	// Try restrictive filtering.
	{
		services := fill()
		filt := newAclFilter(acl.DenyAll(), nil, false)
		filt.filterNodeServices(&services)
		if len((*services).Services) != 0 {
			t.Fatalf("bad: %#v", (*services).Services)
		}
	}

	// Allowed to see the service but not the node.
	policy, err := acl.Parse(`
service "foo" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err := acl.New(acl.DenyAll(), policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// This will work because version 8 ACLs aren't being enforced.
	{
		services := fill()
		filt := newAclFilter(perms, nil, false)
		filt.filterNodeServices(&services)
		if len((*services).Services) != 1 {
			t.Fatalf("bad: %#v", (*services).Services)
		}
	}

	// But with version 8 the node will block it.
	{
		services := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterNodeServices(&services)
		if services != nil {
			t.Fatalf("bad: %#v", services)
		}
	}

	// Chain on access to the node.
	policy, err = acl.Parse(`
node "node1" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now it should go through.
	{
		services := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterNodeServices(&services)
		if len((*services).Services) != 1 {
			t.Fatalf("bad: %#v", (*services).Services)
		}
	}
}

func TestACL_filterCheckServiceNodes(t *testing.T) {
	// Create some nodes.
	fill := func() structs.CheckServiceNodes {
		return structs.CheckServiceNodes{
			structs.CheckServiceNode{
				Node: &structs.Node{
					Node: "node1",
				},
				Service: &structs.NodeService{
					ID:      "foo",
					Service: "foo",
				},
				Checks: structs.HealthChecks{
					&structs.HealthCheck{
						Node:        "node1",
						CheckID:     "check1",
						ServiceName: "foo",
					},
				},
			},
		}
	}

	// Try permissive filtering.
	{
		nodes := fill()
		filt := newAclFilter(acl.AllowAll(), nil, false)
		filt.filterCheckServiceNodes(&nodes)
		if len(nodes) != 1 {
			t.Fatalf("bad: %#v", nodes)
		}
		if len(nodes[0].Checks) != 1 {
			t.Fatalf("bad: %#v", nodes[0].Checks)
		}
	}

	// Try restrictive filtering.
	{
		nodes := fill()
		filt := newAclFilter(acl.DenyAll(), nil, false)
		filt.filterCheckServiceNodes(&nodes)
		if len(nodes) != 0 {
			t.Fatalf("bad: %#v", nodes)
		}
	}

	// Allowed to see the service but not the node.
	policy, err := acl.Parse(`
service "foo" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err := acl.New(acl.DenyAll(), policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// This will work because version 8 ACLs aren't being enforced.
	{
		nodes := fill()
		filt := newAclFilter(perms, nil, false)
		filt.filterCheckServiceNodes(&nodes)
		if len(nodes) != 1 {
			t.Fatalf("bad: %#v", nodes)
		}
		if len(nodes[0].Checks) != 1 {
			t.Fatalf("bad: %#v", nodes[0].Checks)
		}
	}

	// But with version 8 the node will block it.
	{
		nodes := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterCheckServiceNodes(&nodes)
		if len(nodes) != 0 {
			t.Fatalf("bad: %#v", nodes)
		}
	}

	// Chain on access to the node.
	policy, err = acl.Parse(`
node "node1" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now it should go through.
	{
		nodes := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterCheckServiceNodes(&nodes)
		if len(nodes) != 1 {
			t.Fatalf("bad: %#v", nodes)
		}
		if len(nodes[0].Checks) != 1 {
			t.Fatalf("bad: %#v", nodes[0].Checks)
		}
	}
}

func TestACL_filterCoordinates(t *testing.T) {
	// Create some coordinates.
	coords := structs.Coordinates{
		&structs.Coordinate{
			Node:  "node1",
			Coord: generateRandomCoordinate(),
		},
		&structs.Coordinate{
			Node:  "node2",
			Coord: generateRandomCoordinate(),
		},
	}

	// Try permissive filtering.
	filt := newAclFilter(acl.AllowAll(), nil, false)
	filt.filterCoordinates(&coords)
	if len(coords) != 2 {
		t.Fatalf("bad: %#v", coords)
	}

	// Try restrictive filtering without version 8 ACL enforcement.
	filt = newAclFilter(acl.DenyAll(), nil, false)
	filt.filterCoordinates(&coords)
	if len(coords) != 2 {
		t.Fatalf("bad: %#v", coords)
	}

	// Try restrictive filtering with version 8 ACL enforcement.
	filt = newAclFilter(acl.DenyAll(), nil, true)
	filt.filterCoordinates(&coords)
	if len(coords) != 0 {
		t.Fatalf("bad: %#v", coords)
	}
}

func TestACL_filterSessions(t *testing.T) {
	// Create a session list.
	sessions := structs.Sessions{
		&structs.Session{
			Node: "foo",
		},
		&structs.Session{
			Node: "bar",
		},
	}

	// Try permissive filtering.
	filt := newAclFilter(acl.AllowAll(), nil, true)
	filt.filterSessions(&sessions)
	if len(sessions) != 2 {
		t.Fatalf("bad: %#v", sessions)
	}

	// Try restrictive filtering but with version 8 enforcement turned off.
	filt = newAclFilter(acl.DenyAll(), nil, false)
	filt.filterSessions(&sessions)
	if len(sessions) != 2 {
		t.Fatalf("bad: %#v", sessions)
	}

	// Try restrictive filtering with version 8 enforcement turned on.
	filt = newAclFilter(acl.DenyAll(), nil, true)
	filt.filterSessions(&sessions)
	if len(sessions) != 0 {
		t.Fatalf("bad: %#v", sessions)
	}
}

func TestACL_filterNodeDump(t *testing.T) {
	// Create a node dump.
	fill := func() structs.NodeDump {
		return structs.NodeDump{
			&structs.NodeInfo{
				Node: "node1",
				Services: []*structs.NodeService{
					&structs.NodeService{
						ID:      "foo",
						Service: "foo",
					},
				},
				Checks: []*structs.HealthCheck{
					&structs.HealthCheck{
						Node:        "node1",
						CheckID:     "check1",
						ServiceName: "foo",
					},
				},
			},
		}
	}

	// Try permissive filtering.
	{
		dump := fill()
		filt := newAclFilter(acl.AllowAll(), nil, false)
		filt.filterNodeDump(&dump)
		if len(dump) != 1 {
			t.Fatalf("bad: %#v", dump)
		}
		if len(dump[0].Services) != 1 {
			t.Fatalf("bad: %#v", dump[0].Services)
		}
		if len(dump[0].Checks) != 1 {
			t.Fatalf("bad: %#v", dump[0].Checks)
		}
	}

	// Try restrictive filtering.
	{
		dump := fill()
		filt := newAclFilter(acl.DenyAll(), nil, false)
		filt.filterNodeDump(&dump)
		if len(dump) != 1 {
			t.Fatalf("bad: %#v", dump)
		}
		if len(dump[0].Services) != 0 {
			t.Fatalf("bad: %#v", dump[0].Services)
		}
		if len(dump[0].Checks) != 0 {
			t.Fatalf("bad: %#v", dump[0].Checks)
		}
	}

	// Allowed to see the service but not the node.
	policy, err := acl.Parse(`
service "foo" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err := acl.New(acl.DenyAll(), policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// This will work because version 8 ACLs aren't being enforced.
	{
		dump := fill()
		filt := newAclFilter(perms, nil, false)
		filt.filterNodeDump(&dump)
		if len(dump) != 1 {
			t.Fatalf("bad: %#v", dump)
		}
		if len(dump[0].Services) != 1 {
			t.Fatalf("bad: %#v", dump[0].Services)
		}
		if len(dump[0].Checks) != 1 {
			t.Fatalf("bad: %#v", dump[0].Checks)
		}
	}

	// But with version 8 the node will block it.
	{
		dump := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterNodeDump(&dump)
		if len(dump) != 0 {
			t.Fatalf("bad: %#v", dump)
		}
	}

	// Chain on access to the node.
	policy, err = acl.Parse(`
node "node1" {
  policy = "read"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now it should go through.
	{
		dump := fill()
		filt := newAclFilter(perms, nil, true)
		filt.filterNodeDump(&dump)
		if len(dump) != 1 {
			t.Fatalf("bad: %#v", dump)
		}
		if len(dump[0].Services) != 1 {
			t.Fatalf("bad: %#v", dump[0].Services)
		}
		if len(dump[0].Checks) != 1 {
			t.Fatalf("bad: %#v", dump[0].Checks)
		}
	}
}

func TestACL_filterNodes(t *testing.T) {
	// Create a nodes list.
	nodes := structs.Nodes{
		&structs.Node{
			Node: "foo",
		},
		&structs.Node{
			Node: "bar",
		},
	}

	// Try permissive filtering.
	filt := newAclFilter(acl.AllowAll(), nil, true)
	filt.filterNodes(&nodes)
	if len(nodes) != 2 {
		t.Fatalf("bad: %#v", nodes)
	}

	// Try restrictive filtering but with version 8 enforcement turned off.
	filt = newAclFilter(acl.DenyAll(), nil, false)
	filt.filterNodes(&nodes)
	if len(nodes) != 2 {
		t.Fatalf("bad: %#v", nodes)
	}

	// Try restrictive filtering with version 8 enforcement turned on.
	filt = newAclFilter(acl.DenyAll(), nil, true)
	filt.filterNodes(&nodes)
	if len(nodes) != 0 {
		t.Fatalf("bad: %#v", nodes)
	}
}

func TestACL_redactPreparedQueryTokens(t *testing.T) {
	query := &structs.PreparedQuery{
		ID:    "f004177f-2c28-83b7-4229-eacc25fe55d1",
		Token: "root",
	}

	expected := &structs.PreparedQuery{
		ID:    "f004177f-2c28-83b7-4229-eacc25fe55d1",
		Token: "root",
	}

	// Try permissive filtering with a management token. This will allow the
	// embedded token to be seen.
	filt := newAclFilter(acl.ManageAll(), nil, false)
	filt.redactPreparedQueryTokens(&query)
	if !reflect.DeepEqual(query, expected) {
		t.Fatalf("bad: %#v", &query)
	}

	// Hang on to the entry with a token, which needs to survive the next
	// operation.
	original := query

	// Now try permissive filtering with a client token, which should cause
	// the embedded token to get redacted.
	filt = newAclFilter(acl.AllowAll(), nil, false)
	filt.redactPreparedQueryTokens(&query)
	expected.Token = redactedToken
	if !reflect.DeepEqual(query, expected) {
		t.Fatalf("bad: %#v", *query)
	}

	// Make sure that the original object didn't lose its token.
	if original.Token != "root" {
		t.Fatalf("bad token: %s", original.Token)
	}
}

func TestACL_filterPreparedQueries(t *testing.T) {
	queries := structs.PreparedQueries{
		&structs.PreparedQuery{
			ID: "f004177f-2c28-83b7-4229-eacc25fe55d1",
		},
		&structs.PreparedQuery{
			ID:   "f004177f-2c28-83b7-4229-eacc25fe55d2",
			Name: "query-with-no-token",
		},
		&structs.PreparedQuery{
			ID:    "f004177f-2c28-83b7-4229-eacc25fe55d3",
			Name:  "query-with-a-token",
			Token: "root",
		},
	}

	expected := structs.PreparedQueries{
		&structs.PreparedQuery{
			ID: "f004177f-2c28-83b7-4229-eacc25fe55d1",
		},
		&structs.PreparedQuery{
			ID:   "f004177f-2c28-83b7-4229-eacc25fe55d2",
			Name: "query-with-no-token",
		},
		&structs.PreparedQuery{
			ID:    "f004177f-2c28-83b7-4229-eacc25fe55d3",
			Name:  "query-with-a-token",
			Token: "root",
		},
	}

	// Try permissive filtering with a management token. This will allow the
	// embedded token to be seen.
	filt := newAclFilter(acl.ManageAll(), nil, false)
	filt.filterPreparedQueries(&queries)
	if !reflect.DeepEqual(queries, expected) {
		t.Fatalf("bad: %#v", queries)
	}

	// Hang on to the entry with a token, which needs to survive the next
	// operation.
	original := queries[2]

	// Now try permissive filtering with a client token, which should cause
	// the embedded token to get redacted, and the query with no name to get
	// filtered out.
	filt = newAclFilter(acl.AllowAll(), nil, false)
	filt.filterPreparedQueries(&queries)
	expected[2].Token = redactedToken
	expected = append(structs.PreparedQueries{}, expected[1], expected[2])
	if !reflect.DeepEqual(queries, expected) {
		t.Fatalf("bad: %#v", queries)
	}

	// Make sure that the original object didn't lose its token.
	if original.Token != "root" {
		t.Fatalf("bad token: %s", original.Token)
	}

	// Now try restrictive filtering.
	filt = newAclFilter(acl.DenyAll(), nil, false)
	filt.filterPreparedQueries(&queries)
	if len(queries) != 0 {
		t.Fatalf("bad: %#v", queries)
	}
}

func TestACL_unhandledFilterType(t *testing.T) {
	defer func(t *testing.T) {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}(t)

	// Create the server
	dir, token, srv, client := testACLFilterServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer client.Close()

	// Pass an unhandled type into the ACL filter.
	srv.filterACL(token, &structs.HealthCheck{})
}

func TestACL_vetRegisterWithACL(t *testing.T) {
	args := &structs.RegisterRequest{
		Node:    "nope",
		Address: "127.0.0.1",
	}

	// With a nil ACL, the update should be allowed.
	if err := vetRegisterWithACL(nil, args, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create a basic node policy.
	policy, err := acl.Parse(`
node "node" {
  policy = "write"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err := acl.New(acl.DenyAll(), policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// With that policy, the update should now be blocked for node reasons.
	err = vetRegisterWithACL(perms, args, nil)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Now use a permitted node name.
	args.Node = "node"
	if err := vetRegisterWithACL(perms, args, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Build some node info that matches what we have now.
	ns := &structs.NodeServices{
		Node: &structs.Node{
			Node:    "node",
			Address: "127.0.0.1",
		},
		Services: make(map[string]*structs.NodeService),
	}

	// Try to register a service, which should be blocked.
	args.Service = &structs.NodeService{
		Service: "service",
		ID:      "my-id",
	}
	err = vetRegisterWithACL(perms, args, ns)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Chain on a basic service policy.
	policy, err = acl.Parse(`
service "service" {
  policy = "write"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// With the service ACL, the update should go through.
	if err := vetRegisterWithACL(perms, args, ns); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Add an existing service that they are clobbering and aren't allowed
	// to write to.
	ns.Services["my-id"] = &structs.NodeService{
		Service: "other",
		ID:      "my-id",
	}
	err = vetRegisterWithACL(perms, args, ns)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Chain on a policy that allows them to write to the other service.
	policy, err = acl.Parse(`
service "other" {
  policy = "write"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now it should go through.
	if err := vetRegisterWithACL(perms, args, ns); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Try creating the node and the service at once by having no existing
	// node record. This should be ok since we have node and service
	// permissions.
	if err := vetRegisterWithACL(perms, args, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Add a node-level check to the member, which should be rejected.
	args.Check = &structs.HealthCheck{
		Node: "node",
	}
	err = vetRegisterWithACL(perms, args, ns)
	if err == nil || !strings.Contains(err.Error(), "check member must be nil") {
		t.Fatalf("bad: %v", err)
	}

	// Move the check into the slice, but give a bad node name.
	args.Check.Node = "nope"
	args.Checks = append(args.Checks, args.Check)
	args.Check = nil
	err = vetRegisterWithACL(perms, args, ns)
	if err == nil || !strings.Contains(err.Error(), "doesn't match register request node") {
		t.Fatalf("bad: %v", err)
	}

	// Fix the node name, which should now go through.
	args.Checks[0].Node = "node"
	if err := vetRegisterWithACL(perms, args, ns); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Add a service-level check.
	args.Checks = append(args.Checks, &structs.HealthCheck{
		Node:      "node",
		ServiceID: "my-id",
	})
	if err := vetRegisterWithACL(perms, args, ns); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Try creating everything at once. This should be ok since we have all
	// the permissions we need. It also makes sure that we can register a
	// new node, service, and associated checks.
	if err := vetRegisterWithACL(perms, args, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Nil out the service registration, which'll skip the special case
	// and force us to look at the ns data (it will look like we are
	// writing to the "other" service which also has "my-id").
	args.Service = nil
	if err := vetRegisterWithACL(perms, args, ns); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Chain on a policy that forbids them to write to the other service.
	policy, err = acl.Parse(`
service "other" {
  policy = "deny"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// This should get rejected.
	err = vetRegisterWithACL(perms, args, ns)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Change the existing service data to point to a service name they
	// car write to. This should go through.
	ns.Services["my-id"] = &structs.NodeService{
		Service: "service",
		ID:      "my-id",
	}
	if err := vetRegisterWithACL(perms, args, ns); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Chain on a policy that forbids them to write to the node.
	policy, err = acl.Parse(`
node "node" {
  policy = "deny"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err = acl.New(perms, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// This should get rejected because there's a node-level check in here.
	err = vetRegisterWithACL(perms, args, ns)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Change the node-level check into a service check, and then it should
	// go through.
	args.Checks[0].ServiceID = "my-id"
	if err := vetRegisterWithACL(perms, args, ns); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Finally, attempt to update the node part of the data and make sure
	// that gets rejected since they no longer have permissions.
	args.Address = "127.0.0.2"
	err = vetRegisterWithACL(perms, args, ns)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}
}

func TestACL_vetDeregisterWithACL(t *testing.T) {
	args := &structs.DeregisterRequest{
		Node: "nope",
	}

	// With a nil ACL, the update should be allowed.
	if err := vetDeregisterWithACL(nil, args, nil, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create a basic node policy.
	policy, err := acl.Parse(`
node "node" {
  policy = "write"
}
service "service" {
  policy = "write"
}
`)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	perms, err := acl.New(acl.DenyAll(), policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// With that policy, the update should now be blocked for node reasons.
	err = vetDeregisterWithACL(perms, args, nil, nil)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Now use a permitted node name.
	args.Node = "node"
	if err := vetDeregisterWithACL(perms, args, nil, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Try an unknown check.
	args.CheckID = "check-id"
	err = vetDeregisterWithACL(perms, args, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "Unknown check") {
		t.Fatalf("bad: %v", err)
	}

	// Now pass in a check that should be blocked.
	nc := &structs.HealthCheck{
		Node:        "node",
		CheckID:     "check-id",
		ServiceID:   "service-id",
		ServiceName: "nope",
	}
	err = vetDeregisterWithACL(perms, args, nil, nc)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Change it to an allowed service, which should go through.
	nc.ServiceName = "service"
	if err := vetDeregisterWithACL(perms, args, nil, nc); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Switch to a node check that should be blocked.
	args.Node = "nope"
	nc.Node = "nope"
	nc.ServiceID = ""
	nc.ServiceName = ""
	err = vetDeregisterWithACL(perms, args, nil, nc)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Switch to an allowed node check, which should go through.
	args.Node = "node"
	nc.Node = "node"
	if err := vetDeregisterWithACL(perms, args, nil, nc); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Try an unknown service.
	args.ServiceID = "service-id"
	err = vetDeregisterWithACL(perms, args, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "Unknown service") {
		t.Fatalf("bad: %v", err)
	}

	// Now pass in a service that should be blocked.
	ns := &structs.NodeService{
		ID:      "service-id",
		Service: "nope",
	}
	err = vetDeregisterWithACL(perms, args, ns, nil)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("bad: %v", err)
	}

	// Change it to an allowed service, which should go through.
	ns.Service = "service"
	if err := vetDeregisterWithACL(perms, args, ns, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
}
