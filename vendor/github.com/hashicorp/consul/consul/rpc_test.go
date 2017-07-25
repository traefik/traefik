package consul

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/state"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/net-rpc-msgpackrpc"
)

func TestRPC_NoLeader_Fail(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.RPCHoldTimeout = 1 * time.Millisecond
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
	}
	var out struct{}

	// Make sure we eventually fail with a no leader error, which we should
	// see given the short timeout.
	err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out)
	if err.Error() != structs.ErrNoLeader.Error() {
		t.Fatalf("bad: %v", err)
	}

	// Now make sure it goes through.
	testutil.WaitForLeader(t, s1.RPC, "dc1")
	err = msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out)
	if err != nil {
		t.Fatalf("bad: %v", err)
	}
}

func TestRPC_NoLeader_Retry(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.RPCHoldTimeout = 10 * time.Second
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
	}
	var out struct{}

	// This isn't sure-fire but tries to check that we don't have a
	// leader going into the RPC, so we exercise the retry logic.
	if ok, _ := s1.getLeader(); ok {
		t.Fatalf("should not have a leader yet")
	}

	// The timeout is long enough to ride out any reasonable leader
	// election.
	err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out)
	if err != nil {
		t.Fatalf("bad: %v", err)
	}
}

func TestRPC_blockingQuery(t *testing.T) {
	dir, s := testServer(t)
	defer os.RemoveAll(dir)
	defer s.Shutdown()

	// Perform a non-blocking query.
	{
		var opts structs.QueryOptions
		var meta structs.QueryMeta
		var calls int
		fn := func(ws memdb.WatchSet, state *state.StateStore) error {
			calls++
			return nil
		}
		if err := s.blockingQuery(&opts, &meta, fn); err != nil {
			t.Fatalf("err: %v", err)
		}
		if calls != 1 {
			t.Fatalf("bad: %d", calls)
		}
	}

	// Perform a blocking query that gets woken up and loops around once.
	{
		opts := structs.QueryOptions{
			MinQueryIndex: 3,
		}
		var meta structs.QueryMeta
		var calls int
		fn := func(ws memdb.WatchSet, state *state.StateStore) error {
			if calls == 0 {
				meta.Index = 3

				fakeCh := make(chan struct{})
				close(fakeCh)
				ws.Add(fakeCh)
			} else {
				meta.Index = 4
			}
			calls++
			return nil
		}
		if err := s.blockingQuery(&opts, &meta, fn); err != nil {
			t.Fatalf("err: %v", err)
		}
		if calls != 2 {
			t.Fatalf("bad: %d", calls)
		}
	}

	// Perform a query that blocks and gets interrupted when the state store
	// is abandoned.
	{
		opts := structs.QueryOptions{
			MinQueryIndex: 3,
		}
		var meta structs.QueryMeta
		var calls int
		fn := func(ws memdb.WatchSet, state *state.StateStore) error {
			if calls == 0 {
				meta.Index = 3

				snap, err := s.fsm.Snapshot()
				if err != nil {
					t.Fatalf("err: %v", err)
				}
				defer snap.Release()

				buf := bytes.NewBuffer(nil)
				sink := &MockSink{buf, false}
				if err := snap.Persist(sink); err != nil {
					t.Fatalf("err: %v", err)
				}

				if err := s.fsm.Restore(sink); err != nil {
					t.Fatalf("err: %v", err)
				}
			}
			calls++
			return nil
		}
		if err := s.blockingQuery(&opts, &meta, fn); err != nil {
			t.Fatalf("err: %v", err)
		}
		if calls != 1 {
			t.Fatalf("bad: %d", calls)
		}
	}
}
