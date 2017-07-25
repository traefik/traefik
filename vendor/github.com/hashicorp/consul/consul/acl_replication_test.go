package consul

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/testutil"
)

func TestACLReplication_Sorter(t *testing.T) {
	acls := structs.ACLs{
		&structs.ACL{ID: "a"},
		&structs.ACL{ID: "b"},
		&structs.ACL{ID: "c"},
	}

	sorter := &aclIterator{acls, 0}
	if len := sorter.Len(); len != 3 {
		t.Fatalf("bad: %d", len)
	}
	if !sorter.Less(0, 1) {
		t.Fatalf("should be less")
	}
	if sorter.Less(1, 0) {
		t.Fatalf("should not be less")
	}
	if !sort.IsSorted(sorter) {
		t.Fatalf("should be sorted")
	}

	expected := structs.ACLs{
		&structs.ACL{ID: "b"},
		&structs.ACL{ID: "a"},
		&structs.ACL{ID: "c"},
	}
	sorter.Swap(0, 1)
	if !reflect.DeepEqual(acls, expected) {
		t.Fatalf("bad: %v", acls)
	}
	if sort.IsSorted(sorter) {
		t.Fatalf("should not be sorted")
	}
	sort.Sort(sorter)
	if !sort.IsSorted(sorter) {
		t.Fatalf("should be sorted")
	}
}

func TestACLReplication_Iterator(t *testing.T) {
	acls := structs.ACLs{}

	iter := newACLIterator(acls)
	if front := iter.Front(); front != nil {
		t.Fatalf("bad: %v", front)
	}
	iter.Next()
	if front := iter.Front(); front != nil {
		t.Fatalf("bad: %v", front)
	}

	acls = structs.ACLs{
		&structs.ACL{ID: "a"},
		&structs.ACL{ID: "b"},
		&structs.ACL{ID: "c"},
	}
	iter = newACLIterator(acls)
	if front := iter.Front(); front != acls[0] {
		t.Fatalf("bad: %v", front)
	}
	iter.Next()
	if front := iter.Front(); front != acls[1] {
		t.Fatalf("bad: %v", front)
	}
	iter.Next()
	if front := iter.Front(); front != acls[2] {
		t.Fatalf("bad: %v", front)
	}
	iter.Next()
	if front := iter.Front(); front != nil {
		t.Fatalf("bad: %v", front)
	}
}

func TestACLReplication_reconcileACLs(t *testing.T) {
	parseACLs := func(raw string) structs.ACLs {
		var acls structs.ACLs
		for _, key := range strings.Split(raw, "|") {
			if len(key) == 0 {
				continue
			}

			tuple := strings.Split(key, ":")
			index, err := strconv.Atoi(tuple[1])
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			acl := &structs.ACL{
				ID:    tuple[0],
				Rules: tuple[2],
				RaftIndex: structs.RaftIndex{
					ModifyIndex: uint64(index),
				},
			}
			acls = append(acls, acl)
		}
		return acls
	}

	parseChanges := func(changes structs.ACLRequests) string {
		var ret string
		for i, change := range changes {
			if i > 0 {
				ret += "|"
			}
			ret += fmt.Sprintf("%s:%s:%s", change.Op, change.ACL.ID, change.ACL.Rules)
		}
		return ret
	}

	tests := []struct {
		local           string
		remote          string
		lastRemoteIndex uint64
		expected        string
	}{
		// Everything empty.
		{
			local:           "",
			remote:          "",
			lastRemoteIndex: 0,
			expected:        "",
		},
		// First time with empty local.
		{
			local:           "",
			remote:          "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			lastRemoteIndex: 0,
			expected:        "set:bbb:X|set:ccc:X|set:ddd:X|set:eee:X",
		},
		// Remote not sorted.
		{
			local:           "",
			remote:          "ddd:2:X|bbb:3:X|ccc:9:X|eee:11:X",
			lastRemoteIndex: 0,
			expected:        "set:bbb:X|set:ccc:X|set:ddd:X|set:eee:X",
		},
		// Neither side sorted.
		{
			local:           "ddd:2:X|bbb:3:X|ccc:9:X|eee:11:X",
			remote:          "ccc:9:X|bbb:3:X|ddd:2:X|eee:11:X",
			lastRemoteIndex: 0,
			expected:        "",
		},
		// Fully replicated, nothing to do.
		{
			local:           "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			remote:          "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			lastRemoteIndex: 0,
			expected:        "",
		},
		// Change an ACL.
		{
			local:           "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			remote:          "bbb:3:X|ccc:33:Y|ddd:2:X|eee:11:X",
			lastRemoteIndex: 0,
			expected:        "set:ccc:Y",
		},
		// Change an ACL, but mask the change by the last replicated
		// index. This isn't how things work normally, but it proves
		// we are skipping the full compare based on the index.
		{
			local:           "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			remote:          "bbb:3:X|ccc:33:Y|ddd:2:X|eee:11:X",
			lastRemoteIndex: 33,
			expected:        "",
		},
		// Empty everything out.
		{
			local:           "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			remote:          "",
			lastRemoteIndex: 0,
			expected:        "delete:bbb:X|delete:ccc:X|delete:ddd:X|delete:eee:X",
		},
		// Adds on the ends and in the middle.
		{
			local:           "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			remote:          "aaa:99:X|bbb:3:X|ccc:9:X|ccx:101:X|ddd:2:X|eee:11:X|fff:102:X",
			lastRemoteIndex: 0,
			expected:        "set:aaa:X|set:ccx:X|set:fff:X",
		},
		// Deletes on the ends and in the middle.
		{
			local:           "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			remote:          "ccc:9:X",
			lastRemoteIndex: 0,
			expected:        "delete:bbb:X|delete:ddd:X|delete:eee:X",
		},
		// Everything.
		{
			local:           "bbb:3:X|ccc:9:X|ddd:2:X|eee:11:X",
			remote:          "aaa:99:X|bbb:3:X|ccx:101:X|ddd:103:Y|eee:11:X|fff:102:X",
			lastRemoteIndex: 11,
			expected:        "set:aaa:X|delete:ccc:X|set:ccx:X|set:ddd:Y|set:fff:X",
		},
	}
	for i, test := range tests {
		local, remote := parseACLs(test.local), parseACLs(test.remote)
		changes := reconcileACLs(local, remote, test.lastRemoteIndex)
		if actual := parseChanges(changes); actual != test.expected {
			t.Errorf("test case %d failed: %s", i, actual)
		}
	}
}

func TestACLReplication_updateLocalACLs_RateLimit(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc2"
		c.ACLDatacenter = "dc1"
		c.ACLReplicationToken = "secret"
		c.ACLReplicationApplyLimit = 1
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	testutil.WaitForLeader(t, s1.RPC, "dc2")

	changes := structs.ACLRequests{
		&structs.ACLRequest{
			Op: structs.ACLSet,
			ACL: structs.ACL{
				ID:   "secret",
				Type: "client",
			},
		},
	}

	// Should be throttled to 1 Hz.
	start := time.Now()
	if err := s1.updateLocalACLs(changes); err != nil {
		t.Fatalf("err: %v", err)
	}
	if dur := time.Now().Sub(start); dur < time.Second {
		t.Fatalf("too slow: %9.6f", dur.Seconds())
	}

	changes = append(changes,
		&structs.ACLRequest{
			Op: structs.ACLSet,
			ACL: structs.ACL{
				ID:   "secret",
				Type: "client",
			},
		})

	// Should be throttled to 1 Hz.
	start = time.Now()
	if err := s1.updateLocalACLs(changes); err != nil {
		t.Fatalf("err: %v", err)
	}
	if dur := time.Now().Sub(start); dur < 2*time.Second {
		t.Fatalf("too fast: %9.6f", dur.Seconds())
	}
}

func TestACLReplication_IsACLReplicationEnabled(t *testing.T) {
	// ACLs not enabled.
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = ""
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	if s1.IsACLReplicationEnabled() {
		t.Fatalf("should not be enabled")
	}

	// ACLs enabled but not replication.
	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc2"
		c.ACLDatacenter = "dc1"
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()
	if s2.IsACLReplicationEnabled() {
		t.Fatalf("should not be enabled")
	}

	// ACLs enabled with replication.
	dir3, s3 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc2"
		c.ACLDatacenter = "dc1"
		c.ACLReplicationToken = "secret"
	})
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()
	if !s3.IsACLReplicationEnabled() {
		t.Fatalf("should be enabled")
	}

	// ACLs enabled and replication token set, but inside the ACL datacenter
	// so replication should be disabled.
	dir4, s4 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc1"
		c.ACLDatacenter = "dc1"
		c.ACLReplicationToken = "secret"
	})
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()
	if s4.IsACLReplicationEnabled() {
		t.Fatalf("should not be enabled")
	}
}

func TestACLReplication(t *testing.T) {
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
		c.ACLReplicationToken = "root"
		c.ACLReplicationInterval = 0
		c.ACLReplicationApplyLimit = 1000000
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	// Try to join.
	addr := fmt.Sprintf("127.0.0.1:%d",
		s1.config.SerfWANConfig.MemberlistConfig.BindPort)
	if _, err := s2.JoinWAN([]string{addr}); err != nil {
		t.Fatalf("err: %v", err)
	}
	testutil.WaitForLeader(t, s1.RPC, "dc1")
	testutil.WaitForLeader(t, s1.RPC, "dc2")

	// Create a bunch of new tokens.
	var id string
	for i := 0; i < 1000; i++ {
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
		if err := s1.RPC("ACL.Apply", &arg, &id); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	checkSame := func() (bool, error) {
		index, remote, err := s1.fsm.State().ACLList(nil)
		if err != nil {
			return false, err
		}
		_, local, err := s2.fsm.State().ACLList(nil)
		if err != nil {
			return false, err
		}
		if len(remote) != len(local) {
			return false, nil
		}
		for i, acl := range remote {
			if !acl.IsSame(local[i]) {
				return false, nil
			}
		}

		var status structs.ACLReplicationStatus
		s2.aclReplicationStatusLock.RLock()
		status = s2.aclReplicationStatus
		s2.aclReplicationStatusLock.RUnlock()
		if !status.Enabled || !status.Running ||
			status.ReplicatedIndex != index ||
			status.SourceDatacenter != "dc1" {
			return false, nil
		}

		return true, nil
	}

	// Wait for the replica to converge.
	if err := testutil.WaitForResult(checkSame); err != nil {
		t.Fatalf("ACLs didn't converge")
	}

	// Create more new tokens.
	for i := 0; i < 1000; i++ {
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
		var dontCare string
		if err := s1.RPC("ACL.Apply", &arg, &dontCare); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Wait for the replica to converge.
	if err := testutil.WaitForResult(checkSame); err != nil {
		t.Fatalf("ACLs didn't converge")
	}

	// Delete a token.
	arg := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLDelete,
		ACL: structs.ACL{
			ID: id,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var dontCare string
	if err := s1.RPC("ACL.Apply", &arg, &dontCare); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait for the replica to converge.
	if err := testutil.WaitForResult(checkSame); err != nil {
		t.Fatalf("ACLs didn't converge")
	}
}
