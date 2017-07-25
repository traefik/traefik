package state

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-memdb"
)

func TestStateStore_GC(t *testing.T) {
	// Build up a fast GC.
	ttl := 10 * time.Millisecond
	gran := 5 * time.Millisecond
	gc, err := NewTombstoneGC(ttl, gran)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Enable it and attach it to the state store.
	gc.SetEnabled(true)
	s, err := NewStateStore(gc)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Create some KV pairs.
	testSetKey(t, s, 1, "foo", "foo")
	testSetKey(t, s, 2, "foo/bar", "bar")
	testSetKey(t, s, 3, "foo/baz", "bar")
	testSetKey(t, s, 4, "foo/moo", "bar")
	testSetKey(t, s, 5, "foo/zoo", "bar")

	// Delete a key and make sure the GC sees it.
	if err := s.KVSDelete(6, "foo/zoo"); err != nil {
		t.Fatalf("err: %s", err)
	}
	select {
	case idx := <-gc.ExpireCh():
		if idx != 6 {
			t.Fatalf("bad index: %d", idx)
		}
	case <-time.After(2 * ttl):
		t.Fatalf("GC never fired")
	}

	// Check for the same behavior with a tree delete.
	if err := s.KVSDeleteTree(7, "foo/moo"); err != nil {
		t.Fatalf("err: %s", err)
	}
	select {
	case idx := <-gc.ExpireCh():
		if idx != 7 {
			t.Fatalf("bad index: %d", idx)
		}
	case <-time.After(2 * ttl):
		t.Fatalf("GC never fired")
	}

	// Check for the same behavior with a CAS delete.
	if ok, err := s.KVSDeleteCAS(8, 3, "foo/baz"); !ok || err != nil {
		t.Fatalf("err: %s", err)
	}
	select {
	case idx := <-gc.ExpireCh():
		if idx != 8 {
			t.Fatalf("bad index: %d", idx)
		}
	case <-time.After(2 * ttl):
		t.Fatalf("GC never fired")
	}

	// Finally, try it with an expiring session.
	testRegisterNode(t, s, 9, "node1")
	session := &structs.Session{
		ID:       testUUID(),
		Node:     "node1",
		Behavior: structs.SessionKeysDelete,
	}
	if err := s.SessionCreate(10, session); err != nil {
		t.Fatalf("err: %s", err)
	}
	d := &structs.DirEntry{
		Key:     "lock",
		Session: session.ID,
	}
	if ok, err := s.KVSLock(11, d); !ok || err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.SessionDestroy(12, session.ID); err != nil {
		t.Fatalf("err: %s", err)
	}
	select {
	case idx := <-gc.ExpireCh():
		if idx != 12 {
			t.Fatalf("bad index: %d", idx)
		}
	case <-time.After(2 * ttl):
		t.Fatalf("GC never fired")
	}
}

func TestStateStore_ReapTombstones(t *testing.T) {
	s := testStateStore(t)

	// Create some KV pairs.
	testSetKey(t, s, 1, "foo", "foo")
	testSetKey(t, s, 2, "foo/bar", "bar")
	testSetKey(t, s, 3, "foo/baz", "bar")
	testSetKey(t, s, 4, "foo/moo", "bar")
	testSetKey(t, s, 5, "foo/zoo", "bar")

	// Call a delete on some specific keys.
	if err := s.KVSDelete(6, "foo/baz"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := s.KVSDelete(7, "foo/moo"); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Pull out the list and check the index, which should come from the
	// tombstones.
	idx, _, err := s.KVSList(nil, "foo/")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}

	// Reap the tombstones <= 6.
	if err := s.ReapTombstones(6); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Should still be good because 7 is in there.
	idx, _, err = s.KVSList(nil, "foo/")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now reap them all.
	if err := s.ReapTombstones(7); err != nil {
		t.Fatalf("err: %s", err)
	}

	// At this point the sub index will slide backwards.
	idx, _, err = s.KVSList(nil, "foo/")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// Make sure the tombstones are actually gone.
	snap := s.Snapshot()
	defer snap.Close()
	stones, err := snap.Tombstones()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if stones.Next() != nil {
		t.Fatalf("unexpected extra tombstones")
	}
}

func TestStateStore_KVSSet_KVSGet(t *testing.T) {
	s := testStateStore(t)

	// Get on an nonexistent key returns nil.
	ws := memdb.NewWatchSet()
	idx, result, err := s.KVSGet(ws, "foo")
	if result != nil || err != nil || idx != 0 {
		t.Fatalf("expected (0, nil, nil), got : (%#v, %#v, %#v)", idx, result, err)
	}

	// Write a new K/V entry to the store.
	entry := &structs.DirEntry{
		Key:   "foo",
		Value: []byte("bar"),
	}
	if err := s.KVSSet(1, entry); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Retrieve the K/V entry again.
	ws = memdb.NewWatchSet()
	idx, result, err = s.KVSGet(ws, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result == nil {
		t.Fatalf("expected k/v pair, got nothing")
	}
	if idx != 1 {
		t.Fatalf("bad index: %d", idx)
	}

	// Check that the index was injected into the result.
	if result.CreateIndex != 1 || result.ModifyIndex != 1 {
		t.Fatalf("bad index: %d, %d", result.CreateIndex, result.ModifyIndex)
	}

	// Check that the value matches.
	if v := string(result.Value); v != "bar" {
		t.Fatalf("expected 'bar', got: '%s'", v)
	}

	// Updating the entry works and changes the index.
	update := &structs.DirEntry{
		Key:   "foo",
		Value: []byte("baz"),
	}
	if err := s.KVSSet(2, update); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Fetch the kv pair and check.
	ws = memdb.NewWatchSet()
	idx, result, err = s.KVSGet(ws, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.CreateIndex != 1 || result.ModifyIndex != 2 {
		t.Fatalf("bad index: %d, %d", result.CreateIndex, result.ModifyIndex)
	}
	if v := string(result.Value); v != "baz" {
		t.Fatalf("expected 'baz', got '%s'", v)
	}
	if idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}

	// Attempt to set the session during an update.
	update = &structs.DirEntry{
		Key:     "foo",
		Value:   []byte("zoo"),
		Session: "nope",
	}
	if err := s.KVSSet(3, update); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Fetch the kv pair and check.
	ws = memdb.NewWatchSet()
	idx, result, err = s.KVSGet(ws, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.CreateIndex != 1 || result.ModifyIndex != 3 {
		t.Fatalf("bad index: %d, %d", result.CreateIndex, result.ModifyIndex)
	}
	if v := string(result.Value); v != "zoo" {
		t.Fatalf("expected 'zoo', got '%s'", v)
	}
	if result.Session != "" {
		t.Fatalf("expected empty session, got '%s", result.Session)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Make a real session and then lock the key to set the session.
	testRegisterNode(t, s, 4, "node1")
	session := testUUID()
	if err := s.SessionCreate(5, &structs.Session{ID: session, Node: "node1"}); err != nil {
		t.Fatalf("err: %s", err)
	}
	update = &structs.DirEntry{
		Key:     "foo",
		Value:   []byte("locked"),
		Session: session,
	}
	ok, err := s.KVSLock(6, update)
	if !ok || err != nil {
		t.Fatalf("didn't get the lock: %v %s", ok, err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Fetch the kv pair and check.
	ws = memdb.NewWatchSet()
	idx, result, err = s.KVSGet(ws, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.CreateIndex != 1 || result.ModifyIndex != 6 {
		t.Fatalf("bad index: %d, %d", result.CreateIndex, result.ModifyIndex)
	}
	if v := string(result.Value); v != "locked" {
		t.Fatalf("expected 'zoo', got '%s'", v)
	}
	if result.Session != session {
		t.Fatalf("expected session, got '%s", result.Session)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now make an update without the session and make sure it gets applied
	// and doesn't take away the session (it is allowed to change the value).
	update = &structs.DirEntry{
		Key:   "foo",
		Value: []byte("stoleit"),
	}
	if err := s.KVSSet(7, update); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Fetch the kv pair and check.
	ws = memdb.NewWatchSet()
	idx, result, err = s.KVSGet(ws, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.CreateIndex != 1 || result.ModifyIndex != 7 {
		t.Fatalf("bad index: %d, %d", result.CreateIndex, result.ModifyIndex)
	}
	if v := string(result.Value); v != "stoleit" {
		t.Fatalf("expected 'zoo', got '%s'", v)
	}
	if result.Session != session {
		t.Fatalf("expected session, got '%s", result.Session)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}

	// Setting some unrelated key should not fire the watch.
	testSetKey(t, s, 8, "other", "yup")
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Fetch a key that doesn't exist and make sure we get the right
	// response.
	idx, result, err = s.KVSGet(nil, "nope")
	if result != nil || err != nil || idx != 8 {
		t.Fatalf("expected (8, nil, nil), got : (%#v, %#v, %#v)", idx, result, err)
	}
}

func TestStateStore_KVSList(t *testing.T) {
	s := testStateStore(t)

	// Listing an empty KVS returns nothing
	ws := memdb.NewWatchSet()
	idx, entries, err := s.KVSList(ws, "")
	if idx != 0 || entries != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, entries, err)
	}

	// Create some KVS entries
	testSetKey(t, s, 1, "foo", "foo")
	testSetKey(t, s, 2, "foo/bar", "bar")
	testSetKey(t, s, 3, "foo/bar/zip", "zip")
	testSetKey(t, s, 4, "foo/bar/zip/zorp", "zorp")
	testSetKey(t, s, 5, "foo/bar/baz", "baz")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// List out all of the keys
	idx, entries, err = s.KVSList(nil, "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// Check that all of the keys were returned
	if n := len(entries); n != 5 {
		t.Fatalf("expected 5 kvs entries, got: %d", n)
	}

	// Try listing with a provided prefix
	idx, entries, err = s.KVSList(nil, "foo/bar/zip")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Check that only the keys in the prefix were returned
	if n := len(entries); n != 2 {
		t.Fatalf("expected 2 kvs entries, got: %d", n)
	}
	if entries[0].Key != "foo/bar/zip" || entries[1].Key != "foo/bar/zip/zorp" {
		t.Fatalf("bad: %#v", entries)
	}

	// Delete a key and make sure the index comes from the tombstone.
	ws = memdb.NewWatchSet()
	idx, _, err = s.KVSList(ws, "foo/bar/baz")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := s.KVSDelete(6, "foo/bar/baz"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
	ws = memdb.NewWatchSet()
	idx, _, err = s.KVSList(ws, "foo/bar/baz")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Set a different key to bump the index. This shouldn't fire the
	// watch since there's a different prefix.
	testSetKey(t, s, 7, "some/other/key", "")
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Make sure we get the right index from the tombstone.
	idx, _, err = s.KVSList(nil, "foo/bar/baz")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now reap the tombstones and make sure we get the latest index
	// since there are no matching keys.
	if err := s.ReapTombstones(6); err != nil {
		t.Fatalf("err: %s", err)
	}
	idx, _, err = s.KVSList(nil, "foo/bar/baz")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}

	// List all the keys to make sure the index is also correct.
	idx, _, err = s.KVSList(nil, "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVSListKeys(t *testing.T) {
	s := testStateStore(t)

	// Listing keys with no results returns nil.
	ws := memdb.NewWatchSet()
	idx, keys, err := s.KVSListKeys(ws, "", "")
	if idx != 0 || keys != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, keys, err)
	}

	// Create some keys.
	testSetKey(t, s, 1, "foo", "foo")
	testSetKey(t, s, 2, "foo/bar", "bar")
	testSetKey(t, s, 3, "foo/bar/baz", "baz")
	testSetKey(t, s, 4, "foo/bar/zip", "zip")
	testSetKey(t, s, 5, "foo/bar/zip/zam", "zam")
	testSetKey(t, s, 6, "foo/bar/zip/zorp", "zorp")
	testSetKey(t, s, 7, "some/other/prefix", "nack")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// List all the keys.
	idx, keys, err = s.KVSListKeys(nil, "", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(keys) != 7 {
		t.Fatalf("bad keys: %#v", keys)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}

	// Query using a prefix and pass a separator.
	idx, keys, err = s.KVSListKeys(nil, "foo/bar/", "/")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(keys) != 3 {
		t.Fatalf("bad keys: %#v", keys)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Subset of the keys was returned.
	expect := []string{"foo/bar/baz", "foo/bar/zip", "foo/bar/zip/"}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatalf("bad keys: %#v", keys)
	}

	// Listing keys with no separator returns everything.
	idx, keys, err = s.KVSListKeys(nil, "foo", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}
	expect = []string{"foo", "foo/bar", "foo/bar/baz", "foo/bar/zip",
		"foo/bar/zip/zam", "foo/bar/zip/zorp"}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatalf("bad keys: %#v", keys)
	}

	// Delete a key and make sure the index comes from the tombstone.
	ws = memdb.NewWatchSet()
	idx, _, err = s.KVSListKeys(ws, "foo/bar/baz", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := s.KVSDelete(8, "foo/bar/baz"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
	ws = memdb.NewWatchSet()
	idx, _, err = s.KVSListKeys(ws, "foo/bar/baz", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 8 {
		t.Fatalf("bad index: %d", idx)
	}

	// Set a different key to bump the index. This shouldn't fire the watch
	// since there's a different prefix.
	testSetKey(t, s, 9, "some/other/key", "")
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Make sure the index still comes from the tombstone.
	idx, _, err = s.KVSListKeys(nil, "foo/bar/baz", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 8 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now reap the tombstones and make sure we get the latest index
	// since there are no matching keys.
	if err := s.ReapTombstones(8); err != nil {
		t.Fatalf("err: %s", err)
	}
	idx, _, err = s.KVSListKeys(nil, "foo/bar/baz", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}

	// List all the keys to make sure the index is also correct.
	idx, _, err = s.KVSListKeys(nil, "", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVSDelete(t *testing.T) {
	s := testStateStore(t)

	// Create some KV pairs
	testSetKey(t, s, 1, "foo", "foo")
	testSetKey(t, s, 2, "foo/bar", "bar")

	// Call a delete on a specific key
	if err := s.KVSDelete(3, "foo"); err != nil {
		t.Fatalf("err: %s", err)
	}

	// The entry was removed from the state store
	tx := s.db.Txn(false)
	defer tx.Abort()
	e, err := tx.First("kvs", "id", "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if e != nil {
		t.Fatalf("expected kvs entry to be deleted, got: %#v", e)
	}

	// Try fetching the other keys to ensure they still exist
	e, err = tx.First("kvs", "id", "foo/bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if e == nil || string(e.(*structs.DirEntry).Value) != "bar" {
		t.Fatalf("bad kvs entry: %#v", e)
	}

	// Check that the index table was updated
	if idx := s.maxIndex("kvs"); idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Check that the tombstone was created and that prevents the index
	// from sliding backwards.
	idx, _, err := s.KVSList(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now reap the tombstone and watch the index revert to the remaining
	// foo/bar key's index.
	if err := s.ReapTombstones(3); err != nil {
		t.Fatalf("err: %s", err)
	}
	idx, _, err = s.KVSList(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}

	// Deleting a nonexistent key should be idempotent and not return an
	// error
	if err := s.KVSDelete(4, "foo"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx := s.maxIndex("kvs"); idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVSDeleteCAS(t *testing.T) {
	s := testStateStore(t)

	// Create some KV entries
	testSetKey(t, s, 1, "foo", "foo")
	testSetKey(t, s, 2, "bar", "bar")
	testSetKey(t, s, 3, "baz", "baz")

	// Do a CAS delete with an index lower than the entry
	ok, err := s.KVSDeleteCAS(4, 1, "bar")
	if ok || err != nil {
		t.Fatalf("expected (false, nil), got: (%v, %#v)", ok, err)
	}

	// Check that the index is untouched and the entry
	// has not been deleted.
	idx, e, err := s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if e == nil {
		t.Fatalf("expected a kvs entry, got nil")
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Do another CAS delete, this time with the correct index
	// which should cause the delete to take place.
	ok, err = s.KVSDeleteCAS(4, 2, "bar")
	if !ok || err != nil {
		t.Fatalf("expected (true, nil), got: (%v, %#v)", ok, err)
	}

	// Entry was deleted and index was updated
	idx, e, err = s.KVSGet(nil, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if e != nil {
		t.Fatalf("entry should be deleted")
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Add another key to bump the index.
	testSetKey(t, s, 5, "some/other/key", "baz")

	// Check that the tombstone was created and that prevents the index
	// from sliding backwards.
	idx, _, err = s.KVSList(nil, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now reap the tombstone and watch the index move up to the table
	// index since there are no matching keys.
	if err := s.ReapTombstones(4); err != nil {
		t.Fatalf("err: %s", err)
	}
	idx, _, err = s.KVSList(nil, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// A delete on a nonexistent key should be idempotent and not return an
	// error
	ok, err = s.KVSDeleteCAS(6, 2, "bar")
	if !ok || err != nil {
		t.Fatalf("expected (true, nil), got: (%v, %#v)", ok, err)
	}
	if idx := s.maxIndex("kvs"); idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVSSetCAS(t *testing.T) {
	s := testStateStore(t)

	// Doing a CAS with ModifyIndex != 0 and no existing entry
	// is a no-op.
	entry := &structs.DirEntry{
		Key:   "foo",
		Value: []byte("foo"),
		RaftIndex: structs.RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 1,
		},
	}
	ok, err := s.KVSSetCAS(2, entry)
	if ok || err != nil {
		t.Fatalf("expected (false, nil), got: (%#v, %#v)", ok, err)
	}

	// Check that nothing was actually stored
	tx := s.db.Txn(false)
	if e, err := tx.First("kvs", "id", "foo"); e != nil || err != nil {
		t.Fatalf("expected (nil, nil), got: (%#v, %#v)", e, err)
	}
	tx.Abort()

	// Index was not updated
	if idx := s.maxIndex("kvs"); idx != 0 {
		t.Fatalf("bad index: %d", idx)
	}

	// Doing a CAS with a ModifyIndex of zero when no entry exists
	// performs the set and saves into the state store.
	entry = &structs.DirEntry{
		Key:   "foo",
		Value: []byte("foo"),
		RaftIndex: structs.RaftIndex{
			CreateIndex: 0,
			ModifyIndex: 0,
		},
	}
	ok, err = s.KVSSetCAS(2, entry)
	if !ok || err != nil {
		t.Fatalf("expected (true, nil), got: (%#v, %#v)", ok, err)
	}

	// Entry was inserted
	idx, entry, err := s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if string(entry.Value) != "foo" || entry.CreateIndex != 2 || entry.ModifyIndex != 2 {
		t.Fatalf("bad entry: %#v", entry)
	}
	if idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}

	// Doing a CAS with a ModifyIndex of zero when an entry exists does
	// not do anything.
	entry = &structs.DirEntry{
		Key:   "foo",
		Value: []byte("foo"),
		RaftIndex: structs.RaftIndex{
			CreateIndex: 0,
			ModifyIndex: 0,
		},
	}
	ok, err = s.KVSSetCAS(3, entry)
	if ok || err != nil {
		t.Fatalf("expected (false, nil), got: (%#v, %#v)", ok, err)
	}

	// Doing a CAS with a ModifyIndex which does not match the current
	// index does not do anything.
	entry = &structs.DirEntry{
		Key:   "foo",
		Value: []byte("bar"),
		RaftIndex: structs.RaftIndex{
			CreateIndex: 3,
			ModifyIndex: 3,
		},
	}
	ok, err = s.KVSSetCAS(3, entry)
	if ok || err != nil {
		t.Fatalf("expected (false, nil), got: (%#v, %#v)", ok, err)
	}

	// Entry was not updated in the store
	idx, entry, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if string(entry.Value) != "foo" || entry.CreateIndex != 2 || entry.ModifyIndex != 2 {
		t.Fatalf("bad entry: %#v", entry)
	}
	if idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}

	// Doing a CAS with the proper current index should make the
	// modification.
	entry = &structs.DirEntry{
		Key:   "foo",
		Value: []byte("bar"),
		RaftIndex: structs.RaftIndex{
			CreateIndex: 2,
			ModifyIndex: 2,
		},
	}
	ok, err = s.KVSSetCAS(3, entry)
	if !ok || err != nil {
		t.Fatalf("expected (true, nil), got: (%#v, %#v)", ok, err)
	}

	// Entry was updated
	idx, entry, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if string(entry.Value) != "bar" || entry.CreateIndex != 2 || entry.ModifyIndex != 3 {
		t.Fatalf("bad entry: %#v", entry)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Attempt to update the session during the CAS.
	entry = &structs.DirEntry{
		Key:     "foo",
		Value:   []byte("zoo"),
		Session: "nope",
		RaftIndex: structs.RaftIndex{
			CreateIndex: 2,
			ModifyIndex: 3,
		},
	}
	ok, err = s.KVSSetCAS(4, entry)
	if !ok || err != nil {
		t.Fatalf("expected (true, nil), got: (%#v, %#v)", ok, err)
	}

	// Entry was updated, but the session should have been ignored.
	idx, entry, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if string(entry.Value) != "zoo" || entry.CreateIndex != 2 || entry.ModifyIndex != 4 ||
		entry.Session != "" {
		t.Fatalf("bad entry: %#v", entry)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now lock it and try the update, which should keep the session.
	testRegisterNode(t, s, 5, "node1")
	session := testUUID()
	if err := s.SessionCreate(6, &structs.Session{ID: session, Node: "node1"}); err != nil {
		t.Fatalf("err: %s", err)
	}
	entry = &structs.DirEntry{
		Key:     "foo",
		Value:   []byte("locked"),
		Session: session,
		RaftIndex: structs.RaftIndex{
			CreateIndex: 2,
			ModifyIndex: 4,
		},
	}
	ok, err = s.KVSLock(6, entry)
	if !ok || err != nil {
		t.Fatalf("didn't get the lock: %v %s", ok, err)
	}
	entry = &structs.DirEntry{
		Key:   "foo",
		Value: []byte("locked"),
		RaftIndex: structs.RaftIndex{
			CreateIndex: 2,
			ModifyIndex: 6,
		},
	}
	ok, err = s.KVSSetCAS(7, entry)
	if !ok || err != nil {
		t.Fatalf("expected (true, nil), got: (%#v, %#v)", ok, err)
	}

	// Entry was updated, and the lock status should have stayed the same.
	idx, entry, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if string(entry.Value) != "locked" || entry.CreateIndex != 2 || entry.ModifyIndex != 7 ||
		entry.Session != session {
		t.Fatalf("bad entry: %#v", entry)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVSDeleteTree(t *testing.T) {
	s := testStateStore(t)

	// Create kvs entries in the state store.
	testSetKey(t, s, 1, "foo/bar", "bar")
	testSetKey(t, s, 2, "foo/bar/baz", "baz")
	testSetKey(t, s, 3, "foo/bar/zip", "zip")
	testSetKey(t, s, 4, "foo/zorp", "zorp")

	// Calling tree deletion which affects nothing does not
	// modify the table index.
	if err := s.KVSDeleteTree(9, "bar"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx := s.maxIndex("kvs"); idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Call tree deletion with a nested prefix.
	if err := s.KVSDeleteTree(5, "foo/bar"); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Check that all the matching keys were deleted
	tx := s.db.Txn(false)
	defer tx.Abort()

	entries, err := tx.Get("kvs", "id")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	num := 0
	for entry := entries.Next(); entry != nil; entry = entries.Next() {
		if entry.(*structs.DirEntry).Key != "foo/zorp" {
			t.Fatalf("unexpected kvs entry: %#v", entry)
		}
		num++
	}

	if num != 1 {
		t.Fatalf("expected 1 key, got: %d", num)
	}

	// Index should be updated if modifications are made
	if idx := s.maxIndex("kvs"); idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// Check that the tombstones ware created and that prevents the index
	// from sliding backwards.
	idx, _, err := s.KVSList(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now reap the tombstones and watch the index revert to the remaining
	// foo/zorp key's index.
	if err := s.ReapTombstones(5); err != nil {
		t.Fatalf("err: %s", err)
	}
	idx, _, err = s.KVSList(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVSLockDelay(t *testing.T) {
	s := testStateStore(t)

	// KVSLockDelay is exercised in the lock/unlock and session invalidation
	// cases below, so we just do a basic check on a nonexistent key here.
	expires := s.KVSLockDelay("/not/there")
	if expires.After(time.Now()) {
		t.Fatalf("bad: %v", expires)
	}
}

func TestStateStore_KVSLock(t *testing.T) {
	s := testStateStore(t)

	// Lock with no session should fail.
	ok, err := s.KVSLock(0, &structs.DirEntry{Key: "foo", Value: []byte("foo")})
	if ok || err == nil || !strings.Contains(err.Error(), "missing session") {
		t.Fatalf("didn't detect missing session: %v %s", ok, err)
	}

	// Now try with a bogus session.
	ok, err = s.KVSLock(1, &structs.DirEntry{Key: "foo", Value: []byte("foo"), Session: testUUID()})
	if ok || err == nil || !strings.Contains(err.Error(), "invalid session") {
		t.Fatalf("didn't detect invalid session: %v %s", ok, err)
	}

	// Make a real session.
	testRegisterNode(t, s, 2, "node1")
	session1 := testUUID()
	if err := s.SessionCreate(3, &structs.Session{ID: session1, Node: "node1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Lock and make the key at the same time.
	ok, err = s.KVSLock(4, &structs.DirEntry{Key: "foo", Value: []byte("foo"), Session: session1})
	if !ok || err != nil {
		t.Fatalf("didn't get the lock: %v %s", ok, err)
	}

	// Make sure the indexes got set properly.
	idx, result, err := s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 1 || result.CreateIndex != 4 || result.ModifyIndex != 4 ||
		string(result.Value) != "foo" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Re-locking with the same session should update the value and report
	// success.
	ok, err = s.KVSLock(5, &structs.DirEntry{Key: "foo", Value: []byte("bar"), Session: session1})
	if !ok || err != nil {
		t.Fatalf("didn't handle locking an already-locked key: %v %s", ok, err)
	}

	// Make sure the indexes got set properly, note that the lock index
	// won't go up since we didn't lock it again.
	idx, result, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 1 || result.CreateIndex != 4 || result.ModifyIndex != 5 ||
		string(result.Value) != "bar" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// Unlock and the re-lock.
	ok, err = s.KVSUnlock(6, &structs.DirEntry{Key: "foo", Value: []byte("baz"), Session: session1})
	if !ok || err != nil {
		t.Fatalf("didn't handle unlocking a locked key: %v %s", ok, err)
	}
	ok, err = s.KVSLock(7, &structs.DirEntry{Key: "foo", Value: []byte("zoo"), Session: session1})
	if !ok || err != nil {
		t.Fatalf("didn't get the lock: %v %s", ok, err)
	}

	// Make sure the indexes got set properly.
	idx, result, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 2 || result.CreateIndex != 4 || result.ModifyIndex != 7 ||
		string(result.Value) != "zoo" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}

	// Lock an existing key.
	testSetKey(t, s, 8, "bar", "bar")
	ok, err = s.KVSLock(9, &structs.DirEntry{Key: "bar", Value: []byte("xxx"), Session: session1})
	if !ok || err != nil {
		t.Fatalf("didn't get the lock: %v %s", ok, err)
	}

	// Make sure the indexes got set properly.
	idx, result, err = s.KVSGet(nil, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 1 || result.CreateIndex != 8 || result.ModifyIndex != 9 ||
		string(result.Value) != "xxx" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}

	// Attempting a re-lock with a different session should also fail.
	session2 := testUUID()
	if err := s.SessionCreate(10, &structs.Session{ID: session2, Node: "node1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Re-locking should not return an error, but will report that it didn't
	// get the lock.
	ok, err = s.KVSLock(11, &structs.DirEntry{Key: "bar", Value: []byte("nope"), Session: session2})
	if ok || err != nil {
		t.Fatalf("didn't handle locking an already-locked key: %v %s", ok, err)
	}

	// Make sure the indexes didn't update.
	idx, result, err = s.KVSGet(nil, "bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 1 || result.CreateIndex != 8 || result.ModifyIndex != 9 ||
		string(result.Value) != "xxx" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVSUnlock(t *testing.T) {
	s := testStateStore(t)

	// Unlock with no session should fail.
	ok, err := s.KVSUnlock(0, &structs.DirEntry{Key: "foo", Value: []byte("bar")})
	if ok || err == nil || !strings.Contains(err.Error(), "missing session") {
		t.Fatalf("didn't detect missing session: %v %s", ok, err)
	}

	// Make a real session.
	testRegisterNode(t, s, 1, "node1")
	session1 := testUUID()
	if err := s.SessionCreate(2, &structs.Session{ID: session1, Node: "node1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Unlock with a real session but no key should not return an error, but
	// will report it didn't unlock anything.
	ok, err = s.KVSUnlock(3, &structs.DirEntry{Key: "foo", Value: []byte("bar"), Session: session1})
	if ok || err != nil {
		t.Fatalf("didn't handle unlocking a missing key: %v %s", ok, err)
	}

	// Make a key and unlock it, without it being locked.
	testSetKey(t, s, 4, "foo", "bar")
	ok, err = s.KVSUnlock(5, &structs.DirEntry{Key: "foo", Value: []byte("baz"), Session: session1})
	if ok || err != nil {
		t.Fatalf("didn't handle unlocking a non-locked key: %v %s", ok, err)
	}

	// Make sure the indexes didn't update.
	idx, result, err := s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 0 || result.CreateIndex != 4 || result.ModifyIndex != 4 ||
		string(result.Value) != "bar" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Lock it with the first session.
	ok, err = s.KVSLock(6, &structs.DirEntry{Key: "foo", Value: []byte("bar"), Session: session1})
	if !ok || err != nil {
		t.Fatalf("didn't get the lock: %v %s", ok, err)
	}

	// Attempt an unlock with another session.
	session2 := testUUID()
	if err := s.SessionCreate(7, &structs.Session{ID: session2, Node: "node1"}); err != nil {
		t.Fatalf("err: %s", err)
	}
	ok, err = s.KVSUnlock(8, &structs.DirEntry{Key: "foo", Value: []byte("zoo"), Session: session2})
	if ok || err != nil {
		t.Fatalf("didn't handle unlocking with the wrong session: %v %s", ok, err)
	}

	// Make sure the indexes didn't update.
	idx, result, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 1 || result.CreateIndex != 4 || result.ModifyIndex != 6 ||
		string(result.Value) != "bar" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}

	// Now do the unlock with the correct session.
	ok, err = s.KVSUnlock(9, &structs.DirEntry{Key: "foo", Value: []byte("zoo"), Session: session1})
	if !ok || err != nil {
		t.Fatalf("didn't handle unlocking with the correct session: %v %s", ok, err)
	}

	// Make sure the indexes got set properly.
	idx, result, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 1 || result.CreateIndex != 4 || result.ModifyIndex != 9 ||
		string(result.Value) != "zoo" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}

	// Unlocking again should fail and not change anything.
	ok, err = s.KVSUnlock(10, &structs.DirEntry{Key: "foo", Value: []byte("nope"), Session: session1})
	if ok || err != nil {
		t.Fatalf("didn't handle unlocking with the previous session: %v %s", ok, err)
	}

	// Make sure the indexes didn't update.
	idx, result, err = s.KVSGet(nil, "foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.LockIndex != 1 || result.CreateIndex != 4 || result.ModifyIndex != 9 ||
		string(result.Value) != "zoo" {
		t.Fatalf("bad entry: %#v", result)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_KVS_Snapshot_Restore(t *testing.T) {
	s := testStateStore(t)

	// Build up some entries to seed.
	entries := structs.DirEntries{
		&structs.DirEntry{
			Key:   "aaa",
			Flags: 23,
			Value: []byte("hello"),
		},
		&structs.DirEntry{
			Key:   "bar/a",
			Value: []byte("one"),
		},
		&structs.DirEntry{
			Key:   "bar/b",
			Value: []byte("two"),
		},
		&structs.DirEntry{
			Key:   "bar/c",
			Value: []byte("three"),
		},
	}
	for i, entry := range entries {
		if err := s.KVSSet(uint64(i+1), entry); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Make a node and session so we can test a locked key.
	testRegisterNode(t, s, 5, "node1")
	session := testUUID()
	if err := s.SessionCreate(6, &structs.Session{ID: session, Node: "node1"}); err != nil {
		t.Fatalf("err: %s", err)
	}
	entries[3].Session = session
	if ok, err := s.KVSLock(7, entries[3]); !ok || err != nil {
		t.Fatalf("didn't get the lock: %v %s", ok, err)
	}

	// This is required for the compare later.
	entries[3].LockIndex = 1

	// Snapshot the keys.
	snap := s.Snapshot()
	defer snap.Close()

	// Alter the real state store.
	if err := s.KVSSet(8, &structs.DirEntry{Key: "aaa", Value: []byte("nope")}); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the snapshot.
	if idx := snap.LastIndex(); idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}
	iter, err := snap.KVs()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	var dump structs.DirEntries
	for entry := iter.Next(); entry != nil; entry = iter.Next() {
		dump = append(dump, entry.(*structs.DirEntry))
	}
	if !reflect.DeepEqual(dump, entries) {
		t.Fatalf("bad: %#v", dump)
	}

	// Restore the values into a new state store.
	func() {
		s := testStateStore(t)
		restore := s.Restore()
		for _, entry := range dump {
			if err := restore.KVS(entry); err != nil {
				t.Fatalf("err: %s", err)
			}
		}
		restore.Commit()

		// Read the restored keys back out and verify they match.
		idx, res, err := s.KVSList(nil, "")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 7 {
			t.Fatalf("bad index: %d", idx)
		}
		if !reflect.DeepEqual(res, entries) {
			t.Fatalf("bad: %#v", res)
		}

		// Check that the index was updated.
		if idx := s.maxIndex("kvs"); idx != 7 {
			t.Fatalf("bad index: %d", idx)
		}
	}()
}

func TestStateStore_Tombstone_Snapshot_Restore(t *testing.T) {
	s := testStateStore(t)

	// Insert a key and then delete it to create a tombstone.
	testSetKey(t, s, 1, "foo/bar", "bar")
	testSetKey(t, s, 2, "foo/bar/baz", "bar")
	testSetKey(t, s, 3, "foo/bar/zoo", "bar")
	if err := s.KVSDelete(4, "foo/bar"); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Snapshot the Tombstones.
	snap := s.Snapshot()
	defer snap.Close()

	// Alter the real state store.
	if err := s.ReapTombstones(4); err != nil {
		t.Fatalf("err: %s", err)
	}
	idx, _, err := s.KVSList(nil, "foo/bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Verify the snapshot.
	stones, err := snap.Tombstones()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	var dump []*Tombstone
	for stone := stones.Next(); stone != nil; stone = stones.Next() {
		dump = append(dump, stone.(*Tombstone))
	}
	if len(dump) != 1 {
		t.Fatalf("bad %#v", dump)
	}
	stone := dump[0]
	if stone.Key != "foo/bar" || stone.Index != 4 {
		t.Fatalf("bad: %#v", stone)
	}

	// Restore the values into a new state store.
	func() {
		s := testStateStore(t)
		restore := s.Restore()
		for _, stone := range dump {
			if err := restore.Tombstone(stone); err != nil {
				t.Fatalf("err: %s", err)
			}
		}
		restore.Commit()

		// See if the stone works properly in a list query.
		idx, _, err := s.KVSList(nil, "foo/bar")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 4 {
			t.Fatalf("bad index: %d", idx)
		}

		// Make sure it reaps correctly. We should still get a 4 for
		// the index here because it will be using the last index from
		// the tombstone table.
		if err := s.ReapTombstones(4); err != nil {
			t.Fatalf("err: %s", err)
		}
		idx, _, err = s.KVSList(nil, "foo/bar")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 4 {
			t.Fatalf("bad index: %d", idx)
		}

		// But make sure the tombstone is actually gone.
		snap := s.Snapshot()
		defer snap.Close()
		stones, err := snap.Tombstones()
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if stones.Next() != nil {
			t.Fatalf("unexpected extra tombstones")
		}
	}()
}
