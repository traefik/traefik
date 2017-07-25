package consul

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/net-rpc-msgpackrpc"
)

func TestTxn_Apply(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Do a super basic request. The state store test covers the details so
	// we just need to be sure that the transaction is sent correctly and
	// the results are converted appropriately.
	arg := structs.TxnRequest{
		Datacenter: "dc1",
		Ops: structs.TxnOps{
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSSet,
					DirEnt: structs.DirEntry{
						Key:   "test",
						Flags: 42,
						Value: []byte("test"),
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSGet,
					DirEnt: structs.DirEntry{
						Key: "test",
					},
				},
			},
		},
	}
	var out structs.TxnResponse
	if err := msgpackrpc.CallWithCodec(codec, "Txn.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the state store directly.
	state := s1.fsm.State()
	_, d, err := state.KVSGet(nil, "test")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d == nil {
		t.Fatalf("should not be nil")
	}
	if d.Flags != 42 ||
		!bytes.Equal(d.Value, []byte("test")) {
		t.Fatalf("bad: %v", d)
	}

	// Verify the transaction's return value.
	expected := structs.TxnResponse{
		Results: structs.TxnResults{
			&structs.TxnResult{
				KV: &structs.DirEntry{
					Key:   "test",
					Flags: 42,
					Value: nil,
					RaftIndex: structs.RaftIndex{
						CreateIndex: d.CreateIndex,
						ModifyIndex: d.ModifyIndex,
					},
				},
			},
			&structs.TxnResult{
				KV: &structs.DirEntry{
					Key:   "test",
					Flags: 42,
					Value: []byte("test"),
					RaftIndex: structs.RaftIndex{
						CreateIndex: d.CreateIndex,
						ModifyIndex: d.ModifyIndex,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("bad %v", out)
	}
}

func TestTxn_Apply_ACLDeny(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Put in a key to read back.
	state := s1.fsm.State()
	d := &structs.DirEntry{
		Key:   "nope",
		Value: []byte("hello"),
	}
	if err := state.KVSSet(1, d); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create the ACL.
	var id string
	{
		arg := structs.ACLRequest{
			Datacenter: "dc1",
			Op:         structs.ACLSet,
			ACL: structs.ACL{
				Name:  "User token",
				Type:  structs.ACLTypeClient,
				Rules: testListRules,
			},
			WriteRequest: structs.WriteRequest{Token: "root"},
		}
		if err := msgpackrpc.CallWithCodec(codec, "ACL.Apply", &arg, &id); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Set up a transaction where every operation should get blocked due to
	// ACLs.
	arg := structs.TxnRequest{
		Datacenter: "dc1",
		Ops: structs.TxnOps{
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSSet,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSDelete,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSDeleteCAS,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSDeleteTree,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSCAS,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSLock,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSUnlock,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSGet,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSGetTree,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSCheckSession,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSCheckIndex,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
		},
		WriteRequest: structs.WriteRequest{
			Token: id,
		},
	}
	var out structs.TxnResponse
	if err := msgpackrpc.CallWithCodec(codec, "Txn.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the transaction's return value.
	var expected structs.TxnResponse
	for i, op := range arg.Ops {
		switch op.KV.Verb {
		case structs.KVSGet, structs.KVSGetTree:
			// These get filtered but won't result in an error.

		default:
			expected.Errors = append(expected.Errors, &structs.TxnError{
				OpIndex: i,
				What:    permissionDeniedErr.Error(),
			})
		}
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("bad %v", out)
	}
}

func TestTxn_Apply_LockDelay(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create and invalidate a session with a lock.
	state := s1.fsm.State()
	if err := state.EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	session := &structs.Session{
		ID:        generateUUID(),
		Node:      "foo",
		LockDelay: 50 * time.Millisecond,
	}
	if err := state.SessionCreate(2, session); err != nil {
		t.Fatalf("err: %v", err)
	}
	id := session.ID
	d := &structs.DirEntry{
		Key:     "test",
		Session: id,
	}
	if ok, err := state.KVSLock(3, d); err != nil || !ok {
		t.Fatalf("err: %v", err)
	}
	if err := state.SessionDestroy(4, id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make a new session that is valid.
	if err := state.SessionCreate(5, session); err != nil {
		t.Fatalf("err: %v", err)
	}
	validId := session.ID

	// Make a lock request via an atomic transaction.
	arg := structs.TxnRequest{
		Datacenter: "dc1",
		Ops: structs.TxnOps{
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSLock,
					DirEnt: structs.DirEntry{
						Key:     "test",
						Session: validId,
					},
				},
			},
		},
	}
	{
		var out structs.TxnResponse
		if err := msgpackrpc.CallWithCodec(codec, "Txn.Apply", &arg, &out); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(out.Results) != 0 ||
			len(out.Errors) != 1 ||
			out.Errors[0].OpIndex != 0 ||
			!strings.Contains(out.Errors[0].What, "due to lock delay") {
			t.Fatalf("bad: %v", out)
		}
	}

	// Wait for lock-delay.
	time.Sleep(50 * time.Millisecond)

	// Should acquire.
	{
		var out structs.TxnResponse
		if err := msgpackrpc.CallWithCodec(codec, "Txn.Apply", &arg, &out); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(out.Results) != 1 ||
			len(out.Errors) != 0 ||
			out.Results[0].KV.LockIndex != 2 {
			t.Fatalf("bad: %v", out)
		}
	}
}

func TestTxn_Read(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Put in a key to read back.
	state := s1.fsm.State()
	d := &structs.DirEntry{
		Key:   "test",
		Value: []byte("hello"),
	}
	if err := state.KVSSet(1, d); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Do a super basic request. The state store test covers the details so
	// we just need to be sure that the transaction is sent correctly and
	// the results are converted appropriately.
	arg := structs.TxnReadRequest{
		Datacenter: "dc1",
		Ops: structs.TxnOps{
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSGet,
					DirEnt: structs.DirEntry{
						Key: "test",
					},
				},
			},
		},
	}
	var out structs.TxnReadResponse
	if err := msgpackrpc.CallWithCodec(codec, "Txn.Read", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the transaction's return value.
	expected := structs.TxnReadResponse{
		TxnResponse: structs.TxnResponse{
			Results: structs.TxnResults{
				&structs.TxnResult{
					KV: &structs.DirEntry{
						Key:   "test",
						Value: []byte("hello"),
						RaftIndex: structs.RaftIndex{
							CreateIndex: 1,
							ModifyIndex: 1,
						},
					},
				},
			},
		},
		QueryMeta: structs.QueryMeta{
			KnownLeader: true,
		},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("bad %v", out)
	}
}

func TestTxn_Read_ACLDeny(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Put in a key to read back.
	state := s1.fsm.State()
	d := &structs.DirEntry{
		Key:   "nope",
		Value: []byte("hello"),
	}
	if err := state.KVSSet(1, d); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create the ACL.
	var id string
	{
		arg := structs.ACLRequest{
			Datacenter: "dc1",
			Op:         structs.ACLSet,
			ACL: structs.ACL{
				Name:  "User token",
				Type:  structs.ACLTypeClient,
				Rules: testListRules,
			},
			WriteRequest: structs.WriteRequest{Token: "root"},
		}
		if err := msgpackrpc.CallWithCodec(codec, "ACL.Apply", &arg, &id); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Set up a transaction where every operation should get blocked due to
	// ACLs.
	arg := structs.TxnReadRequest{
		Datacenter: "dc1",
		Ops: structs.TxnOps{
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSGet,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSGetTree,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSCheckSession,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
			&structs.TxnOp{
				KV: &structs.TxnKVOp{
					Verb: structs.KVSCheckIndex,
					DirEnt: structs.DirEntry{
						Key: "nope",
					},
				},
			},
		},
		QueryOptions: structs.QueryOptions{
			Token: id,
		},
	}
	var out structs.TxnReadResponse
	if err := msgpackrpc.CallWithCodec(codec, "Txn.Read", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the transaction's return value.
	expected := structs.TxnReadResponse{
		QueryMeta: structs.QueryMeta{
			KnownLeader: true,
		},
	}
	for i, op := range arg.Ops {
		switch op.KV.Verb {
		case structs.KVSGet, structs.KVSGetTree:
			// These get filtered but won't result in an error.

		default:
			expected.Errors = append(expected.Errors, &structs.TxnError{
				OpIndex: i,
				What:    permissionDeniedErr.Error(),
			})
		}
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("bad %v", out)
	}
}
