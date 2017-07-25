package consul

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/net-rpc-msgpackrpc"
)

func TestSession_Apply(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Just add a node
	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})

	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node: "foo",
			Name: "my-session",
		},
	}
	var out string
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	id := out

	// Verify
	state := s1.fsm.State()
	_, s, err := state.SessionGet(nil, out)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s == nil {
		t.Fatalf("should not be nil")
	}
	if s.Node != "foo" {
		t.Fatalf("bad: %v", s)
	}
	if s.Name != "my-session" {
		t.Fatalf("bad: %v", s)
	}

	// Do a delete
	arg.Op = structs.SessionDestroy
	arg.Session.ID = out
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify
	_, s, err = state.SessionGet(nil, id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s != nil {
		t.Fatalf("bad: %v", s)
	}
}

func TestSession_DeleteApply(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Just add a node
	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})

	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node:     "foo",
			Name:     "my-session",
			Behavior: structs.SessionKeysDelete,
		},
	}
	var out string
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	id := out

	// Verify
	state := s1.fsm.State()
	_, s, err := state.SessionGet(nil, out)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s == nil {
		t.Fatalf("should not be nil")
	}
	if s.Node != "foo" {
		t.Fatalf("bad: %v", s)
	}
	if s.Name != "my-session" {
		t.Fatalf("bad: %v", s)
	}
	if s.Behavior != structs.SessionKeysDelete {
		t.Fatalf("bad: %v", s)
	}

	// Do a delete
	arg.Op = structs.SessionDestroy
	arg.Session.ID = out
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify
	_, s, err = state.SessionGet(nil, id)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s != nil {
		t.Fatalf("bad: %v", s)
	}
}

func TestSession_Apply_ACLDeny(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = false
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create the ACL.
	req := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name: "User token",
			Type: structs.ACLTypeClient,
			Rules: `
session "foo" {
	policy = "write"
}
`,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}

	var token string
	if err := msgpackrpc.CallWithCodec(codec, "ACL.Apply", &req, &token); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Just add a node.
	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})

	// Try to create without a token, which will go through since version 8
	// enforcement isn't enabled.
	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node: "foo",
			Name: "my-session",
		},
	}
	var id1 string
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &id1); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now turn on version 8 enforcement and try again, it should be denied.
	var id2 string
	s1.config.ACLEnforceVersion8 = true
	err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &id2)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("err: %v", err)
	}

	// Now set a token and try again. This should go through.
	arg.Token = token
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &id2); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Do a delete on the first session with version 8 enforcement off and
	// no token. This should go through.
	var out string
	s1.config.ACLEnforceVersion8 = false
	arg.Op = structs.SessionDestroy
	arg.Token = ""
	arg.Session.ID = id1
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Turn on version 8 enforcement and make sure the delete of the second
	// session fails.
	s1.config.ACLEnforceVersion8 = true
	arg.Session.ID = id2
	err = msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("err: %v", err)
	}

	// Now set a token and try again. This should go through.
	arg.Token = token
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestSession_Get(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})
	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node: "foo",
		},
	}
	var out string
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	getR := structs.SessionSpecificRequest{
		Datacenter: "dc1",
		Session:    out,
	}
	var sessions structs.IndexedSessions
	if err := msgpackrpc.CallWithCodec(codec, "Session.Get", &getR, &sessions); err != nil {
		t.Fatalf("err: %v", err)
	}

	if sessions.Index == 0 {
		t.Fatalf("Bad: %v", sessions)
	}
	if len(sessions.Sessions) != 1 {
		t.Fatalf("Bad: %v", sessions)
	}
	s := sessions.Sessions[0]
	if s.ID != out {
		t.Fatalf("bad: %v", s)
	}
}

func TestSession_List(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})
	ids := []string{}
	for i := 0; i < 5; i++ {
		arg := structs.SessionRequest{
			Datacenter: "dc1",
			Op:         structs.SessionCreate,
			Session: structs.Session{
				Node: "foo",
			},
		}
		var out string
		if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
			t.Fatalf("err: %v", err)
		}
		ids = append(ids, out)
	}

	getR := structs.DCSpecificRequest{
		Datacenter: "dc1",
	}
	var sessions structs.IndexedSessions
	if err := msgpackrpc.CallWithCodec(codec, "Session.List", &getR, &sessions); err != nil {
		t.Fatalf("err: %v", err)
	}

	if sessions.Index == 0 {
		t.Fatalf("Bad: %v", sessions)
	}
	if len(sessions.Sessions) != 5 {
		t.Fatalf("Bad: %v", sessions.Sessions)
	}
	for i := 0; i < len(sessions.Sessions); i++ {
		s := sessions.Sessions[i]
		if !lib.StrContains(ids, s.ID) {
			t.Fatalf("bad: %v", s)
		}
		if s.Node != "foo" {
			t.Fatalf("bad: %v", s)
		}
	}
}

func TestSession_Get_List_NodeSessions_ACLFilter(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = false
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create the ACL.
	req := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name: "User token",
			Type: structs.ACLTypeClient,
			Rules: `
session "foo" {
	policy = "read"
}
`,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}

	var token string
	if err := msgpackrpc.CallWithCodec(codec, "ACL.Apply", &req, &token); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create a node and a session.
	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})
	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node: "foo",
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var out string
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Perform all the read operations, which should go through since version
	// 8 ACL enforcement isn't enabled.
	getR := structs.SessionSpecificRequest{
		Datacenter: "dc1",
		Session:    out,
	}
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.Get", &getR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 1 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}
	listR := structs.DCSpecificRequest{
		Datacenter: "dc1",
	}
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.List", &listR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 1 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}
	nodeR := structs.NodeSpecificRequest{
		Datacenter: "dc1",
		Node:       "foo",
	}
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.NodeSessions", &nodeR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 1 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}

	// Now turn on version 8 enforcement and make sure everything is empty.
	s1.config.ACLEnforceVersion8 = true
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.Get", &getR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 0 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}
	{
		var sessions structs.IndexedSessions

		if err := msgpackrpc.CallWithCodec(codec, "Session.List", &listR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 0 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.NodeSessions", &nodeR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 0 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}

	// Finally, supply the token and make sure the reads are allowed.
	getR.Token = token
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.Get", &getR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 1 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}
	listR.Token = token
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.List", &listR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 1 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}
	nodeR.Token = token
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.NodeSessions", &nodeR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 1 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}

	// Try to get a session that doesn't exist to make sure that's handled
	// correctly by the filter (it will get passed a nil slice).
	getR.Session = "adf4238a-882b-9ddc-4a9d-5b6758e4159e"
	{
		var sessions structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.Get", &getR, &sessions); err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(sessions.Sessions) != 0 {
			t.Fatalf("bad: %v", sessions.Sessions)
		}
	}
}

func TestSession_ApplyTimers(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})
	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node: "foo",
			TTL:  "10s",
		},
	}
	var out string
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check the session map
	if _, ok := s1.sessionTimers[out]; !ok {
		t.Fatalf("missing session timer")
	}

	// Destroy the session
	arg.Op = structs.SessionDestroy
	arg.Session.ID = out
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check the session map
	if _, ok := s1.sessionTimers[out]; ok {
		t.Fatalf("session timer exists")
	}
}

func TestSession_Renew(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")
	TTL := "10s" // the minimum allowed ttl
	ttl := 10 * time.Second

	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})
	ids := []string{}
	for i := 0; i < 5; i++ {
		arg := structs.SessionRequest{
			Datacenter: "dc1",
			Op:         structs.SessionCreate,
			Session: structs.Session{
				Node: "foo",
				TTL:  TTL,
			},
		}
		var out string
		if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
			t.Fatalf("err: %v", err)
		}
		ids = append(ids, out)
	}

	// Verify the timer map is setup
	if len(s1.sessionTimers) != 5 {
		t.Fatalf("missing session timers")
	}

	getR := structs.DCSpecificRequest{
		Datacenter: "dc1",
	}

	var sessions structs.IndexedSessions
	if err := msgpackrpc.CallWithCodec(codec, "Session.List", &getR, &sessions); err != nil {
		t.Fatalf("err: %v", err)
	}

	if sessions.Index == 0 {
		t.Fatalf("Bad: %v", sessions)
	}
	if len(sessions.Sessions) != 5 {
		t.Fatalf("Bad: %v", sessions.Sessions)
	}
	for i := 0; i < len(sessions.Sessions); i++ {
		s := sessions.Sessions[i]
		if !lib.StrContains(ids, s.ID) {
			t.Fatalf("bad: %v", s)
		}
		if s.Node != "foo" {
			t.Fatalf("bad: %v", s)
		}
		if s.TTL != TTL {
			t.Fatalf("bad session TTL: %s %v", s.TTL, s)
		}
		t.Logf("Created session '%s'", s.ID)
	}

	// Sleep for time shorter than internal destroy ttl
	time.Sleep(ttl * structs.SessionTTLMultiplier / 2)

	// renew 3 out of 5 sessions
	for i := 0; i < 3; i++ {
		renewR := structs.SessionSpecificRequest{
			Datacenter: "dc1",
			Session:    ids[i],
		}
		var session structs.IndexedSessions
		if err := msgpackrpc.CallWithCodec(codec, "Session.Renew", &renewR, &session); err != nil {
			t.Fatalf("err: %v", err)
		}

		if session.Index == 0 {
			t.Fatalf("Bad: %v", session)
		}
		if len(session.Sessions) != 1 {
			t.Fatalf("Bad: %v", session.Sessions)
		}

		s := session.Sessions[0]
		if !lib.StrContains(ids, s.ID) {
			t.Fatalf("bad: %v", s)
		}
		if s.Node != "foo" {
			t.Fatalf("bad: %v", s)
		}

		t.Logf("Renewed session '%s'", s.ID)
	}

	// now sleep for 2/3 the internal destroy TTL time for renewed sessions
	// which is more than the internal destroy TTL time for the non-renewed sessions
	time.Sleep((ttl * structs.SessionTTLMultiplier) * 2.0 / 3.0)

	var sessionsL1 structs.IndexedSessions
	if err := msgpackrpc.CallWithCodec(codec, "Session.List", &getR, &sessionsL1); err != nil {
		t.Fatalf("err: %v", err)
	}

	if sessionsL1.Index == 0 {
		t.Fatalf("Bad: %v", sessionsL1)
	}

	t.Logf("Expect 2 sessions to be destroyed")

	for i := 0; i < len(sessionsL1.Sessions); i++ {
		s := sessionsL1.Sessions[i]
		if !lib.StrContains(ids, s.ID) {
			t.Fatalf("bad: %v", s)
		}
		if s.Node != "foo" {
			t.Fatalf("bad: %v", s)
		}
		if s.TTL != TTL {
			t.Fatalf("bad: %v", s)
		}
		if i > 2 {
			t.Errorf("session '%s' should be destroyed", s.ID)
		}
	}

	if len(sessionsL1.Sessions) != 3 {
		t.Fatalf("Bad: %v", sessionsL1.Sessions)
	}

	// now sleep again for ttl*2 - no sessions should still be alive
	time.Sleep(ttl * structs.SessionTTLMultiplier)

	var sessionsL2 structs.IndexedSessions
	if err := msgpackrpc.CallWithCodec(codec, "Session.List", &getR, &sessionsL2); err != nil {
		t.Fatalf("err: %v", err)
	}

	if sessionsL2.Index == 0 {
		t.Fatalf("Bad: %v", sessionsL2)
	}
	if len(sessionsL2.Sessions) != 0 {
		for i := 0; i < len(sessionsL2.Sessions); i++ {
			s := sessionsL2.Sessions[i]
			if !lib.StrContains(ids, s.ID) {
				t.Fatalf("bad: %v", s)
			}
			if s.Node != "foo" {
				t.Fatalf("bad: %v", s)
			}
			if s.TTL != TTL {
				t.Fatalf("bad: %v", s)
			}
			t.Errorf("session '%s' should be destroyed", s.ID)
		}

		t.Fatalf("Bad: %v", sessionsL2.Sessions)
	}
}

func TestSession_Renew_ACLDeny(t *testing.T) {
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.ACLDatacenter = "dc1"
		c.ACLMasterToken = "root"
		c.ACLDefaultPolicy = "deny"
		c.ACLEnforceVersion8 = false
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	// Create the ACL.
	req := structs.ACLRequest{
		Datacenter: "dc1",
		Op:         structs.ACLSet,
		ACL: structs.ACL{
			Name: "User token",
			Type: structs.ACLTypeClient,
			Rules: `
session "foo" {
	policy = "write"
}
`,
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}

	var token string
	if err := msgpackrpc.CallWithCodec(codec, "ACL.Apply", &req, &token); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Just add a node.
	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})

	// Create a session. The token won't matter here since we don't have
	// version 8 ACL enforcement on yet.
	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node: "foo",
			Name: "my-session",
		},
	}
	var id string
	if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &id); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Renew without a token should go through without version 8 ACL
	// enforcement.
	renewR := structs.SessionSpecificRequest{
		Datacenter: "dc1",
		Session:    id,
	}
	var session structs.IndexedSessions
	if err := msgpackrpc.CallWithCodec(codec, "Session.Renew", &renewR, &session); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now turn on version 8 enforcement and the renew should be rejected.
	s1.config.ACLEnforceVersion8 = true
	err := msgpackrpc.CallWithCodec(codec, "Session.Renew", &renewR, &session)
	if err == nil || !strings.Contains(err.Error(), permissionDenied) {
		t.Fatalf("err: %v", err)
	}

	// Set the token and it should go through.
	renewR.Token = token
	if err := msgpackrpc.CallWithCodec(codec, "Session.Renew", &renewR, &session); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestSession_NodeSessions(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"})
	s1.fsm.State().EnsureNode(1, &structs.Node{Node: "bar", Address: "127.0.0.1"})
	ids := []string{}
	for i := 0; i < 10; i++ {
		arg := structs.SessionRequest{
			Datacenter: "dc1",
			Op:         structs.SessionCreate,
			Session: structs.Session{
				Node: "bar",
			},
		}
		if i < 5 {
			arg.Session.Node = "foo"
		}
		var out string
		if err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out); err != nil {
			t.Fatalf("err: %v", err)
		}
		if i < 5 {
			ids = append(ids, out)
		}
	}

	getR := structs.NodeSpecificRequest{
		Datacenter: "dc1",
		Node:       "foo",
	}
	var sessions structs.IndexedSessions
	if err := msgpackrpc.CallWithCodec(codec, "Session.NodeSessions", &getR, &sessions); err != nil {
		t.Fatalf("err: %v", err)
	}

	if sessions.Index == 0 {
		t.Fatalf("Bad: %v", sessions)
	}
	if len(sessions.Sessions) != 5 {
		t.Fatalf("Bad: %v", sessions.Sessions)
	}
	for i := 0; i < len(sessions.Sessions); i++ {
		s := sessions.Sessions[i]
		if !lib.StrContains(ids, s.ID) {
			t.Fatalf("bad: %v", s)
		}
		if s.Node != "foo" {
			t.Fatalf("bad: %v", s)
		}
	}
}

func TestSession_Apply_BadTTL(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.SessionRequest{
		Datacenter: "dc1",
		Op:         structs.SessionCreate,
		Session: structs.Session{
			Node: "foo",
			Name: "my-session",
		},
	}

	// Session with illegal TTL
	arg.Session.TTL = "10z"

	var out string
	err := msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "Session TTL '10z' invalid: time: unknown unit z in duration 10z" {
		t.Fatalf("incorrect error message: %s", err.Error())
	}

	// less than SessionTTLMin
	arg.Session.TTL = "5s"

	err = msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "Invalid Session TTL '5000000000', must be between [10s=24h0m0s]" {
		t.Fatalf("incorrect error message: %s", err.Error())
	}

	// more than SessionTTLMax
	arg.Session.TTL = "100000s"

	err = msgpackrpc.CallWithCodec(codec, "Session.Apply", &arg, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "Invalid Session TTL '100000000000000', must be between [10s=24h0m0s]" {
		t.Fatalf("incorrect error message: %s", err.Error())
	}
}
