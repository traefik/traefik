package state

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/go-memdb"
)

func TestStateStore_SessionCreate_SessionGet(t *testing.T) {
	s := testStateStore(t)

	// SessionGet returns nil if the session doesn't exist
	ws := memdb.NewWatchSet()
	idx, session, err := s.SessionGet(ws, testUUID())
	if session != nil || err != nil {
		t.Fatalf("expected (nil, nil), got: (%#v, %#v)", session, err)
	}
	if idx != 0 {
		t.Fatalf("bad index: %d", idx)
	}

	// Registering without a session ID is disallowed
	err = s.SessionCreate(1, &structs.Session{})
	if err != ErrMissingSessionID {
		t.Fatalf("expected %#v, got: %#v", ErrMissingSessionID, err)
	}

	// Invalid session behavior throws error
	sess := &structs.Session{
		ID:       testUUID(),
		Behavior: "nope",
	}
	err = s.SessionCreate(1, sess)
	if err == nil || !strings.Contains(err.Error(), "session behavior") {
		t.Fatalf("expected session behavior error, got: %#v", err)
	}

	// Registering with an unknown node is disallowed
	sess = &structs.Session{ID: testUUID()}
	if err := s.SessionCreate(1, sess); err != ErrMissingNode {
		t.Fatalf("expected %#v, got: %#v", ErrMissingNode, err)
	}

	// None of the errored operations modified the index
	if idx := s.maxIndex("sessions"); idx != 0 {
		t.Fatalf("bad index: %d", idx)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Valid session is able to register
	testRegisterNode(t, s, 1, "node1")
	sess = &structs.Session{
		ID:   testUUID(),
		Node: "node1",
	}
	if err := s.SessionCreate(2, sess); err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx := s.maxIndex("sessions"); idx != 2 {
		t.Fatalf("bad index: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Retrieve the session again
	ws = memdb.NewWatchSet()
	idx, session, err = s.SessionGet(ws, sess.ID)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}

	// Ensure the session looks correct and was assigned the
	// proper default value for session behavior.
	expect := &structs.Session{
		ID:       sess.ID,
		Behavior: structs.SessionKeysRelease,
		Node:     "node1",
		RaftIndex: structs.RaftIndex{
			CreateIndex: 2,
			ModifyIndex: 2,
		},
	}
	if !reflect.DeepEqual(expect, session) {
		t.Fatalf("bad session: %#v", session)
	}

	// Registering with a non-existent check is disallowed
	sess = &structs.Session{
		ID:     testUUID(),
		Node:   "node1",
		Checks: []types.CheckID{"check1"},
	}
	err = s.SessionCreate(3, sess)
	if err == nil || !strings.Contains(err.Error(), "Missing check") {
		t.Fatalf("expected missing check error, got: %#v", err)
	}

	// Registering with a critical check is disallowed
	testRegisterCheck(t, s, 3, "node1", "", "check1", structs.HealthCritical)
	err = s.SessionCreate(4, sess)
	if err == nil || !strings.Contains(err.Error(), structs.HealthCritical) {
		t.Fatalf("expected critical state error, got: %#v", err)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Registering with a healthy check succeeds (doesn't hit the watch since
	// we are looking at the old session).
	testRegisterCheck(t, s, 4, "node1", "", "check1", structs.HealthPassing)
	if err := s.SessionCreate(5, sess); err != nil {
		t.Fatalf("err: %s", err)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Register a session against two checks.
	testRegisterCheck(t, s, 5, "node1", "", "check2", structs.HealthPassing)
	sess2 := &structs.Session{
		ID:     testUUID(),
		Node:   "node1",
		Checks: []types.CheckID{"check1", "check2"},
	}
	if err := s.SessionCreate(6, sess2); err != nil {
		t.Fatalf("err: %s", err)
	}

	tx := s.db.Txn(false)
	defer tx.Abort()

	// Check mappings were inserted
	{
		check, err := tx.First("session_checks", "session", sess.ID)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if check == nil {
			t.Fatalf("missing session check")
		}
		expectCheck := &sessionCheck{
			Node:    "node1",
			CheckID: "check1",
			Session: sess.ID,
		}
		if actual := check.(*sessionCheck); !reflect.DeepEqual(actual, expectCheck) {
			t.Fatalf("expected %#v, got: %#v", expectCheck, actual)
		}
	}
	checks, err := tx.Get("session_checks", "session", sess2.ID)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	for i, check := 0, checks.Next(); check != nil; i, check = i+1, checks.Next() {
		expectCheck := &sessionCheck{
			Node:    "node1",
			CheckID: types.CheckID(fmt.Sprintf("check%d", i+1)),
			Session: sess2.ID,
		}
		if actual := check.(*sessionCheck); !reflect.DeepEqual(actual, expectCheck) {
			t.Fatalf("expected %#v, got: %#v", expectCheck, actual)
		}
	}

	// Pulling a nonexistent session gives the table index.
	idx, session, err = s.SessionGet(nil, testUUID())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if session != nil {
		t.Fatalf("expected not to get a session: %v", session)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TegstStateStore_SessionList(t *testing.T) {
	s := testStateStore(t)

	// Listing when no sessions exist returns nil
	ws := memdb.NewWatchSet()
	idx, res, err := s.SessionList(ws)
	if idx != 0 || res != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Register some nodes
	testRegisterNode(t, s, 1, "node1")
	testRegisterNode(t, s, 2, "node2")
	testRegisterNode(t, s, 3, "node3")

	// Create some sessions in the state store
	sessions := structs.Sessions{
		&structs.Session{
			ID:       testUUID(),
			Node:     "node1",
			Behavior: structs.SessionKeysDelete,
		},
		&structs.Session{
			ID:       testUUID(),
			Node:     "node2",
			Behavior: structs.SessionKeysRelease,
		},
		&structs.Session{
			ID:       testUUID(),
			Node:     "node3",
			Behavior: structs.SessionKeysDelete,
		},
	}
	for i, session := range sessions {
		if err := s.SessionCreate(uint64(4+i), session); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// List out all of the sessions
	idx, sessionList, err := s.SessionList(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}
	if !reflect.DeepEqual(sessionList, sessions) {
		t.Fatalf("bad: %#v", sessions)
	}
}

func TestStateStore_NodeSessions(t *testing.T) {
	s := testStateStore(t)

	// Listing sessions with no results returns nil
	ws := memdb.NewWatchSet()
	idx, res, err := s.NodeSessions(ws, "node1")
	if idx != 0 || res != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Create the nodes
	testRegisterNode(t, s, 1, "node1")
	testRegisterNode(t, s, 2, "node2")

	// Register some sessions with the nodes
	sessions1 := structs.Sessions{
		&structs.Session{
			ID:   testUUID(),
			Node: "node1",
		},
		&structs.Session{
			ID:   testUUID(),
			Node: "node1",
		},
	}
	sessions2 := []*structs.Session{
		&structs.Session{
			ID:   testUUID(),
			Node: "node2",
		},
		&structs.Session{
			ID:   testUUID(),
			Node: "node2",
		},
	}
	for i, sess := range append(sessions1, sessions2...) {
		if err := s.SessionCreate(uint64(3+i), sess); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Query all of the sessions associated with a specific
	// node in the state store.
	ws1 := memdb.NewWatchSet()
	idx, res, err = s.NodeSessions(ws1, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(res) != len(sessions1) {
		t.Fatalf("bad: %#v", res)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	ws2 := memdb.NewWatchSet()
	idx, res, err = s.NodeSessions(ws2, "node2")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(res) != len(sessions2) {
		t.Fatalf("bad: %#v", res)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Destroying a session on node1 should not affect node2's watch.
	if err := s.SessionDestroy(100, sessions1[0].ID); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws1) {
		t.Fatalf("bad")
	}
	if watchFired(ws2) {
		t.Fatalf("bad")
	}
}

func TestStateStore_SessionDestroy(t *testing.T) {
	s := testStateStore(t)

	// Session destroy is idempotent and returns no error
	// if the session doesn't exist.
	if err := s.SessionDestroy(1, testUUID()); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Ensure the index was not updated if nothing was destroyed.
	if idx := s.maxIndex("sessions"); idx != 0 {
		t.Fatalf("bad index: %d", idx)
	}

	// Register a node.
	testRegisterNode(t, s, 1, "node1")

	// Register a new session
	sess := &structs.Session{
		ID:   testUUID(),
		Node: "node1",
	}
	if err := s.SessionCreate(2, sess); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Destroy the session.
	if err := s.SessionDestroy(3, sess.ID); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Check that the index was updated
	if idx := s.maxIndex("sessions"); idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Make sure the session is really gone.
	tx := s.db.Txn(false)
	sessions, err := tx.Get("sessions", "id")
	if err != nil || sessions.Next() != nil {
		t.Fatalf("session should not exist")
	}
	tx.Abort()
}

func TestStateStore_Session_Snapshot_Restore(t *testing.T) {
	s := testStateStore(t)

	// Register some nodes and checks.
	testRegisterNode(t, s, 1, "node1")
	testRegisterNode(t, s, 2, "node2")
	testRegisterNode(t, s, 3, "node3")
	testRegisterCheck(t, s, 4, "node1", "", "check1", structs.HealthPassing)

	// Create some sessions in the state store.
	session1 := testUUID()
	sessions := structs.Sessions{
		&structs.Session{
			ID:       session1,
			Node:     "node1",
			Behavior: structs.SessionKeysDelete,
			Checks:   []types.CheckID{"check1"},
		},
		&structs.Session{
			ID:        testUUID(),
			Node:      "node2",
			Behavior:  structs.SessionKeysRelease,
			LockDelay: 10 * time.Second,
		},
		&structs.Session{
			ID:       testUUID(),
			Node:     "node3",
			Behavior: structs.SessionKeysDelete,
			TTL:      "1.5s",
		},
	}
	for i, session := range sessions {
		if err := s.SessionCreate(uint64(5+i), session); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Snapshot the sessions.
	snap := s.Snapshot()
	defer snap.Close()

	// Alter the real state store.
	if err := s.SessionDestroy(8, session1); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the snapshot.
	if idx := snap.LastIndex(); idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}
	iter, err := snap.Sessions()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	var dump structs.Sessions
	for session := iter.Next(); session != nil; session = iter.Next() {
		sess := session.(*structs.Session)
		dump = append(dump, sess)

		found := false
		for i, _ := range sessions {
			if sess.ID == sessions[i].ID {
				if !reflect.DeepEqual(sess, sessions[i]) {
					t.Fatalf("bad: %#v", sess)
				}
				found = true
			}
		}
		if !found {
			t.Fatalf("bad: %#v", sess)
		}
	}

	// Restore the sessions into a new state store.
	func() {
		s := testStateStore(t)
		restore := s.Restore()
		for _, session := range dump {
			if err := restore.Session(session); err != nil {
				t.Fatalf("err: %s", err)
			}
		}
		restore.Commit()

		// Read the restored sessions back out and verify that they
		// match.
		idx, res, err := s.SessionList(nil)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 7 {
			t.Fatalf("bad index: %d", idx)
		}
		for _, sess := range res {
			found := false
			for i, _ := range sessions {
				if sess.ID == sessions[i].ID {
					if !reflect.DeepEqual(sess, sessions[i]) {
						t.Fatalf("bad: %#v", sess)
					}
					found = true
				}
			}
			if !found {
				t.Fatalf("bad: %#v", sess)
			}
		}

		// Check that the index was updated.
		if idx := s.maxIndex("sessions"); idx != 7 {
			t.Fatalf("bad index: %d", idx)
		}

		// Manually verify that the session check mapping got restored.
		tx := s.db.Txn(false)
		defer tx.Abort()

		check, err := tx.First("session_checks", "session", session1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if check == nil {
			t.Fatalf("missing session check")
		}
		expectCheck := &sessionCheck{
			Node:    "node1",
			CheckID: "check1",
			Session: session1,
		}
		if actual := check.(*sessionCheck); !reflect.DeepEqual(actual, expectCheck) {
			t.Fatalf("expected %#v, got: %#v", expectCheck, actual)
		}
	}()
}

func TestStateStore_Session_Invalidate_DeleteNode(t *testing.T) {
	s := testStateStore(t)

	// Set up our test environment.
	if err := s.EnsureNode(3, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	session := &structs.Session{
		ID:   testUUID(),
		Node: "foo",
	}
	if err := s.SessionCreate(14, session); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Delete the node and make sure the watch fires.
	ws := memdb.NewWatchSet()
	idx, s2, err := s.SessionGet(ws, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.DeleteNode(15, "foo"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Lookup by ID, should be nil.
	idx, s2, err = s.SessionGet(nil, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s2 != nil {
		t.Fatalf("session should be invalidated")
	}
	if idx != 15 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_Session_Invalidate_DeleteService(t *testing.T) {
	s := testStateStore(t)

	// Set up our test environment.
	if err := s.EnsureNode(11, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(12, "foo", &structs.NodeService{ID: "api", Service: "api", Tags: nil, Address: "", Port: 5000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	check := &structs.HealthCheck{
		Node:      "foo",
		CheckID:   "api",
		Name:      "Can connect",
		Status:    structs.HealthPassing,
		ServiceID: "api",
	}
	if err := s.EnsureCheck(13, check); err != nil {
		t.Fatalf("err: %v", err)
	}
	session := &structs.Session{
		ID:     testUUID(),
		Node:   "foo",
		Checks: []types.CheckID{"api"},
	}
	if err := s.SessionCreate(14, session); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Delete the service and make sure the watch fires.
	ws := memdb.NewWatchSet()
	idx, s2, err := s.SessionGet(ws, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.DeleteService(15, "foo", "api"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Lookup by ID, should be nil.
	idx, s2, err = s.SessionGet(nil, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s2 != nil {
		t.Fatalf("session should be invalidated")
	}
	if idx != 15 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_Session_Invalidate_Critical_Check(t *testing.T) {
	s := testStateStore(t)

	// Set up our test environment.
	if err := s.EnsureNode(3, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	check := &structs.HealthCheck{
		Node:    "foo",
		CheckID: "bar",
		Status:  structs.HealthPassing,
	}
	if err := s.EnsureCheck(13, check); err != nil {
		t.Fatalf("err: %v", err)
	}
	session := &structs.Session{
		ID:     testUUID(),
		Node:   "foo",
		Checks: []types.CheckID{"bar"},
	}
	if err := s.SessionCreate(14, session); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Invalidate the check and make sure the watches fire.
	ws := memdb.NewWatchSet()
	idx, s2, err := s.SessionGet(ws, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	check.Status = structs.HealthCritical
	if err := s.EnsureCheck(15, check); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Lookup by ID, should be nil.
	idx, s2, err = s.SessionGet(nil, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s2 != nil {
		t.Fatalf("session should be invalidated")
	}
	if idx != 15 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_Session_Invalidate_DeleteCheck(t *testing.T) {
	s := testStateStore(t)

	// Set up our test environment.
	if err := s.EnsureNode(3, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	check := &structs.HealthCheck{
		Node:    "foo",
		CheckID: "bar",
		Status:  structs.HealthPassing,
	}
	if err := s.EnsureCheck(13, check); err != nil {
		t.Fatalf("err: %v", err)
	}
	session := &structs.Session{
		ID:     testUUID(),
		Node:   "foo",
		Checks: []types.CheckID{"bar"},
	}
	if err := s.SessionCreate(14, session); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Delete the check and make sure the watches fire.
	ws := memdb.NewWatchSet()
	idx, s2, err := s.SessionGet(ws, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.DeleteCheck(15, "foo", "bar"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Lookup by ID, should be nil.
	idx, s2, err = s.SessionGet(nil, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s2 != nil {
		t.Fatalf("session should be invalidated")
	}
	if idx != 15 {
		t.Fatalf("bad index: %d", idx)
	}

	// Manually make sure the session checks mapping is clear.
	tx := s.db.Txn(false)
	mapping, err := tx.First("session_checks", "session", session.ID)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mapping != nil {
		t.Fatalf("unexpected session check")
	}
	tx.Abort()
}

func TestStateStore_Session_Invalidate_Key_Unlock_Behavior(t *testing.T) {
	s := testStateStore(t)

	// Set up our test environment.
	if err := s.EnsureNode(3, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	session := &structs.Session{
		ID:        testUUID(),
		Node:      "foo",
		LockDelay: 50 * time.Millisecond,
	}
	if err := s.SessionCreate(4, session); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Lock a key with the session.
	d := &structs.DirEntry{
		Key:     "/foo",
		Flags:   42,
		Value:   []byte("test"),
		Session: session.ID,
	}
	ok, err := s.KVSLock(5, d)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("unexpected fail")
	}

	// Delete the node and make sure the watches fire.
	ws := memdb.NewWatchSet()
	idx, s2, err := s.SessionGet(ws, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.DeleteNode(6, "foo"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Lookup by ID, should be nil.
	idx, s2, err = s.SessionGet(nil, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s2 != nil {
		t.Fatalf("session should be invalidated")
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Key should be unlocked.
	idx, d2, err := s.KVSGet(nil, "/foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if d2.ModifyIndex != 6 {
		t.Fatalf("bad index: %v", d2.ModifyIndex)
	}
	if d2.LockIndex != 1 {
		t.Fatalf("bad: %v", *d2)
	}
	if d2.Session != "" {
		t.Fatalf("bad: %v", *d2)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Key should have a lock delay.
	expires := s.KVSLockDelay("/foo")
	if expires.Before(time.Now().Add(30 * time.Millisecond)) {
		t.Fatalf("Bad: %v", expires)
	}
}

func TestStateStore_Session_Invalidate_Key_Delete_Behavior(t *testing.T) {
	s := testStateStore(t)

	// Set up our test environment.
	if err := s.EnsureNode(3, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	session := &structs.Session{
		ID:        testUUID(),
		Node:      "foo",
		LockDelay: 50 * time.Millisecond,
		Behavior:  structs.SessionKeysDelete,
	}
	if err := s.SessionCreate(4, session); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Lock a key with the session.
	d := &structs.DirEntry{
		Key:     "/bar",
		Flags:   42,
		Value:   []byte("test"),
		Session: session.ID,
	}
	ok, err := s.KVSLock(5, d)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("unexpected fail")
	}

	// Delete the node and make sure the watches fire.
	ws := memdb.NewWatchSet()
	idx, s2, err := s.SessionGet(ws, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.DeleteNode(6, "foo"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Lookup by ID, should be nil.
	idx, s2, err = s.SessionGet(nil, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s2 != nil {
		t.Fatalf("session should be invalidated")
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Key should be deleted.
	idx, d2, err := s.KVSGet(nil, "/bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if d2 != nil {
		t.Fatalf("unexpected deleted key")
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Key should have a lock delay.
	expires := s.KVSLockDelay("/bar")
	if expires.Before(time.Now().Add(30 * time.Millisecond)) {
		t.Fatalf("Bad: %v", expires)
	}
}

func TestStateStore_Session_Invalidate_PreparedQuery_Delete(t *testing.T) {
	s := testStateStore(t)

	// Set up our test environment.
	testRegisterNode(t, s, 1, "foo")
	testRegisterService(t, s, 2, "foo", "redis")
	session := &structs.Session{
		ID:   testUUID(),
		Node: "foo",
	}
	if err := s.SessionCreate(3, session); err != nil {
		t.Fatalf("err: %v", err)
	}
	query := &structs.PreparedQuery{
		ID:      testUUID(),
		Session: session.ID,
		Service: structs.ServiceQuery{
			Service: "redis",
		},
	}
	if err := s.PreparedQuerySet(4, query); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Invalidate the session and make sure the watches fire.
	ws := memdb.NewWatchSet()
	idx, s2, err := s.SessionGet(ws, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.SessionDestroy(5, session.ID); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Make sure the session is gone.
	idx, s2, err = s.SessionGet(nil, session.ID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s2 != nil {
		t.Fatalf("session should be invalidated")
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// Make sure the query is gone and the index is updated.
	idx, q2, err := s.PreparedQueryGet(nil, query.ID)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}
	if q2 != nil {
		t.Fatalf("bad: %v", q2)
	}
}
