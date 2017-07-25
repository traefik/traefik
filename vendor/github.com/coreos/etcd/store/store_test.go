// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"testing"
	"time"

	etcdErr "github.com/coreos/etcd/error"
	"github.com/coreos/etcd/pkg/testutil"
	"github.com/jonboulle/clockwork"
)

func TestNewStoreWithNamespaces(t *testing.T) {
	s := newStore("/0", "/1")

	_, err := s.Get("/0", false, false)
	testutil.AssertNil(t, err)
	_, err = s.Get("/1", false, false)
	testutil.AssertNil(t, err)
}

// Ensure that the store can retrieve an existing value.
func TestStoreGetValue(t *testing.T) {
	s := newStore()
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	var eidx uint64 = 1
	e, err := s.Get("/foo", false, false)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "get")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertEqual(t, *e.Node.Value, "bar")
}

// Ensure that any TTL <= minExpireTime becomes Permanent
func TestMinExpireTime(t *testing.T) {
	s := newStore()
	fc := clockwork.NewFakeClock()
	s.clock = fc
	// FakeClock starts at 0, so minExpireTime should be far in the future.. but just in case
	testutil.AssertTrue(t, minExpireTime.After(fc.Now()), "minExpireTime should be ahead of FakeClock!")
	s.Create("/foo", false, "Y", false, TTLOptionSet{ExpireTime: fc.Now().Add(3 * time.Second)})
	fc.Advance(5 * time.Second)
	// Ensure it hasn't expired
	s.DeleteExpiredKeys(fc.Now())
	var eidx uint64 = 1
	e, err := s.Get("/foo", true, false)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "get")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
}

// Ensure that the store can recursively retrieve a directory listing.
// Note that hidden files should not be returned.
func TestStoreGetDirectory(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/bar", false, "X", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/_hidden", false, "*", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/baz", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/baz/bat", false, "Y", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/baz/_hidden", false, "*", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/baz/ttl", false, "Y", false, TTLOptionSet{ExpireTime: fc.Now().Add(time.Second * 3)})
	var eidx uint64 = 7
	e, err := s.Get("/foo", true, false)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "get")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertEqual(t, len(e.Node.Nodes), 2)
	var bazNodes NodeExterns
	for _, node := range e.Node.Nodes {
		switch node.Key {
		case "/foo/bar":
			testutil.AssertEqual(t, *node.Value, "X")
			testutil.AssertEqual(t, node.Dir, false)
		case "/foo/baz":
			testutil.AssertEqual(t, node.Dir, true)
			testutil.AssertEqual(t, len(node.Nodes), 2)
			bazNodes = node.Nodes
		default:
			t.Errorf("key = %s, not matched", node.Key)
		}
	}
	for _, node := range bazNodes {
		switch node.Key {
		case "/foo/baz/bat":
			testutil.AssertEqual(t, *node.Value, "Y")
			testutil.AssertEqual(t, node.Dir, false)
		case "/foo/baz/ttl":
			testutil.AssertEqual(t, *node.Value, "Y")
			testutil.AssertEqual(t, node.Dir, false)
			testutil.AssertEqual(t, node.TTL, int64(3))
		default:
			t.Errorf("key = %s, not matched", node.Key)
		}
	}
}

// Ensure that the store can retrieve a directory in sorted order.
func TestStoreGetSorted(t *testing.T) {
	s := newStore()
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/x", false, "0", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/z", false, "0", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/y", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/y/a", false, "0", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/y/b", false, "0", false, TTLOptionSet{ExpireTime: Permanent})
	var eidx uint64 = 6
	e, err := s.Get("/foo", true, true)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)

	var yNodes NodeExterns
	sortedStrings := []string{"/foo/x", "/foo/y", "/foo/z"}
	for i := range e.Node.Nodes {
		node := e.Node.Nodes[i]
		if node.Key != sortedStrings[i] {
			t.Errorf("expect key = %s, got key = %s", sortedStrings[i], node.Key)
		}
		if node.Key == "/foo/y" {
			yNodes = node.Nodes
		}
	}

	sortedStrings = []string{"/foo/y/a", "/foo/y/b"}
	for i := range yNodes {
		node := yNodes[i]
		if node.Key != sortedStrings[i] {
			t.Errorf("expect key = %s, got key = %s", sortedStrings[i], node.Key)
		}
	}
}

func TestSet(t *testing.T) {
	s := newStore()

	// Set /foo=""
	var eidx uint64 = 1
	e, err := s.Set("/foo", false, "", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "set")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertFalse(t, e.Node.Dir)
	testutil.AssertEqual(t, *e.Node.Value, "")
	testutil.AssertNil(t, e.Node.Nodes)
	testutil.AssertNil(t, e.Node.Expiration)
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(1))

	// Set /foo="bar"
	eidx = 2
	e, err = s.Set("/foo", false, "bar", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "set")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertFalse(t, e.Node.Dir)
	testutil.AssertEqual(t, *e.Node.Value, "bar")
	testutil.AssertNil(t, e.Node.Nodes)
	testutil.AssertNil(t, e.Node.Expiration)
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(2))
	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "")
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(1))
	// Set /foo="baz" (for testing prevNode)
	eidx = 3
	e, err = s.Set("/foo", false, "baz", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "set")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertFalse(t, e.Node.Dir)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	testutil.AssertNil(t, e.Node.Nodes)
	testutil.AssertNil(t, e.Node.Expiration)
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(3))
	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(2))

	// Set /dir as a directory
	eidx = 4
	e, err = s.Set("/dir", true, "", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "set")
	testutil.AssertEqual(t, e.Node.Key, "/dir")
	testutil.AssertTrue(t, e.Node.Dir)
	testutil.AssertNil(t, e.Node.Value)
	testutil.AssertNil(t, e.Node.Nodes)
	testutil.AssertNil(t, e.Node.Expiration)
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(4))
}

// Ensure that the store can create a new key if it doesn't already exist.
func TestStoreCreateValue(t *testing.T) {
	s := newStore()
	// Create /foo=bar
	var eidx uint64 = 1
	e, err := s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertFalse(t, e.Node.Dir)
	testutil.AssertEqual(t, *e.Node.Value, "bar")
	testutil.AssertNil(t, e.Node.Nodes)
	testutil.AssertNil(t, e.Node.Expiration)
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(1))

	// Create /empty=""
	eidx = 2
	e, err = s.Create("/empty", false, "", false, TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/empty")
	testutil.AssertFalse(t, e.Node.Dir)
	testutil.AssertEqual(t, *e.Node.Value, "")
	testutil.AssertNil(t, e.Node.Nodes)
	testutil.AssertNil(t, e.Node.Expiration)
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(2))

}

// Ensure that the store can create a new directory if it doesn't already exist.
func TestStoreCreateDirectory(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	e, err := s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertTrue(t, e.Node.Dir)
}

// Ensure that the store fails to create a key if it already exists.
func TestStoreCreateFailsIfExists(t *testing.T) {
	s := newStore()
	// create /foo as dir
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})

	// create /foo as dir again
	e, _err := s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeNodeExist)
	testutil.AssertEqual(t, err.Message, "Key already exists")
	testutil.AssertEqual(t, err.Cause, "/foo")
	testutil.AssertEqual(t, err.Index, uint64(1))
	testutil.AssertNil(t, e)
}

// Ensure that the store can update a key if it already exists.
func TestStoreUpdateValue(t *testing.T) {
	s := newStore()
	// create /foo=bar
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	// update /foo="bzr"
	var eidx uint64 = 2
	e, err := s.Update("/foo", "baz", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "update")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertFalse(t, e.Node.Dir)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(2))
	// check prevNode
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
	testutil.AssertEqual(t, e.PrevNode.TTL, int64(0))
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(1))

	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	testutil.AssertEqual(t, e.EtcdIndex, eidx)

	// update /foo=""
	eidx = 3
	e, err = s.Update("/foo", "", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "update")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertFalse(t, e.Node.Dir)
	testutil.AssertEqual(t, *e.Node.Value, "")
	testutil.AssertEqual(t, e.Node.TTL, int64(0))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(3))
	// check prevNode
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "baz")
	testutil.AssertEqual(t, e.PrevNode.TTL, int64(0))
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(2))

	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, *e.Node.Value, "")
}

// Ensure that the store cannot update a directory.
func TestStoreUpdateFailsIfDirectory(t *testing.T) {
	s := newStore()
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	e, _err := s.Update("/foo", "baz", TTLOptionSet{ExpireTime: Permanent})
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeNotFile)
	testutil.AssertEqual(t, err.Message, "Not a file")
	testutil.AssertEqual(t, err.Cause, "/foo")
	testutil.AssertNil(t, e)
}

// Ensure that the store can update the TTL on a value.
func TestStoreUpdateValueTTL(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	var eidx uint64 = 2
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	_, err := s.Update("/foo", "baz", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond)})
	testutil.AssertNil(t, err)
	e, _ := s.Get("/foo", false, false)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	fc.Advance(600 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	e, err = s.Get("/foo", false, false)
	testutil.AssertNil(t, e)
	testutil.AssertEqual(t, err.(*etcdErr.Error).ErrorCode, etcdErr.EcodeKeyNotFound)
}

// Ensure that the store can update the TTL on a directory.
func TestStoreUpdateDirTTL(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	var eidx uint64 = 3
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	e, err := s.Update("/foo", "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond)})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.Node.Dir, true)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	e, _ = s.Get("/foo/bar", false, false)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	testutil.AssertEqual(t, e.EtcdIndex, eidx)

	fc.Advance(600 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	e, err = s.Get("/foo/bar", false, false)
	testutil.AssertNil(t, e)
	testutil.AssertEqual(t, err.(*etcdErr.Error).ErrorCode, etcdErr.EcodeKeyNotFound)
}

// Ensure that the store can delete a value.
func TestStoreDeleteValue(t *testing.T) {
	s := newStore()
	var eidx uint64 = 2
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, err := s.Delete("/foo", false, false)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "delete")
	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
}

// Ensure that the store can delete a directory if recursive is specified.
func TestStoreDeleteDiretory(t *testing.T) {
	s := newStore()
	// create directory /foo
	var eidx uint64 = 2
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	// delete /foo with dir = true and recursive = false
	// this should succeed, since the directory is empty
	e, err := s.Delete("/foo", true, false)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "delete")
	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, e.PrevNode.Dir, true)

	// create directory /foo and directory /foo/bar
	s.Create("/foo/bar", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	// delete /foo with dir = true and recursive = false
	// this should fail, since the directory is not empty
	_, err = s.Delete("/foo", true, false)
	testutil.AssertNotNil(t, err)

	// delete /foo with dir=false and recursive = true
	// this should succeed, since recursive implies dir=true
	// and recursively delete should be able to delete all
	// items under the given directory
	e, err = s.Delete("/foo", false, true)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.Action, "delete")

}

// Ensure that the store cannot delete a directory if both of recursive
// and dir are not specified.
func TestStoreDeleteDiretoryFailsIfNonRecursiveAndDir(t *testing.T) {
	s := newStore()
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	e, _err := s.Delete("/foo", false, false)
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeNotFile)
	testutil.AssertEqual(t, err.Message, "Not a file")
	testutil.AssertNil(t, e)
}

func TestRootRdOnly(t *testing.T) {
	s := newStore("/0")

	for _, tt := range []string{"/", "/0"} {
		_, err := s.Set(tt, true, "", TTLOptionSet{ExpireTime: Permanent})
		testutil.AssertNotNil(t, err)

		_, err = s.Delete(tt, true, true)
		testutil.AssertNotNil(t, err)

		_, err = s.Create(tt, true, "", false, TTLOptionSet{ExpireTime: Permanent})
		testutil.AssertNotNil(t, err)

		_, err = s.Update(tt, "", TTLOptionSet{ExpireTime: Permanent})
		testutil.AssertNotNil(t, err)

		_, err = s.CompareAndSwap(tt, "", 0, "", TTLOptionSet{ExpireTime: Permanent})
		testutil.AssertNotNil(t, err)
	}
}

func TestStoreCompareAndDeletePrevValue(t *testing.T) {
	s := newStore()
	var eidx uint64 = 2
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, err := s.CompareAndDelete("/foo", "bar", 0)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "compareAndDelete")
	testutil.AssertEqual(t, e.Node.Key, "/foo")

	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(1))
	testutil.AssertEqual(t, e.PrevNode.CreatedIndex, uint64(1))
}

func TestStoreCompareAndDeletePrevValueFailsIfNotMatch(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, _err := s.CompareAndDelete("/foo", "baz", 0)
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeTestFailed)
	testutil.AssertEqual(t, err.Message, "Compare failed")
	testutil.AssertNil(t, e)
	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, *e.Node.Value, "bar")
}

func TestStoreCompareAndDeletePrevIndex(t *testing.T) {
	s := newStore()
	var eidx uint64 = 2
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, err := s.CompareAndDelete("/foo", "", 1)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "compareAndDelete")
	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(1))
	testutil.AssertEqual(t, e.PrevNode.CreatedIndex, uint64(1))
}

func TestStoreCompareAndDeletePrevIndexFailsIfNotMatch(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, _err := s.CompareAndDelete("/foo", "", 100)
	testutil.AssertNotNil(t, _err)
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeTestFailed)
	testutil.AssertEqual(t, err.Message, "Compare failed")
	testutil.AssertNil(t, e)
	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, *e.Node.Value, "bar")
}

// Ensure that the store cannot delete a directory.
func TestStoreCompareAndDeleteDiretoryFail(t *testing.T) {
	s := newStore()
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	_, _err := s.CompareAndDelete("/foo", "", 0)
	testutil.AssertNotNil(t, _err)
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeNotFile)
}

// Ensure that the store can conditionally update a key if it has a previous value.
func TestStoreCompareAndSwapPrevValue(t *testing.T) {
	s := newStore()
	var eidx uint64 = 2
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, err := s.CompareAndSwap("/foo", "bar", 0, "baz", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "compareAndSwap")
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(1))
	testutil.AssertEqual(t, e.PrevNode.CreatedIndex, uint64(1))

	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
}

// Ensure that the store cannot conditionally update a key if it has the wrong previous value.
func TestStoreCompareAndSwapPrevValueFailsIfNotMatch(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, _err := s.CompareAndSwap("/foo", "wrong_value", 0, "baz", TTLOptionSet{ExpireTime: Permanent})
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeTestFailed)
	testutil.AssertEqual(t, err.Message, "Compare failed")
	testutil.AssertNil(t, e)
	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, *e.Node.Value, "bar")
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
}

// Ensure that the store can conditionally update a key if it has a previous index.
func TestStoreCompareAndSwapPrevIndex(t *testing.T) {
	s := newStore()
	var eidx uint64 = 2
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, err := s.CompareAndSwap("/foo", "", 1, "baz", TTLOptionSet{ExpireTime: Permanent})
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "compareAndSwap")
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	// check prevNode
	testutil.AssertNotNil(t, e.PrevNode)
	testutil.AssertEqual(t, e.PrevNode.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
	testutil.AssertEqual(t, e.PrevNode.ModifiedIndex, uint64(1))
	testutil.AssertEqual(t, e.PrevNode.CreatedIndex, uint64(1))

	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
}

// Ensure that the store cannot conditionally update a key if it has the wrong previous index.
func TestStoreCompareAndSwapPrevIndexFailsIfNotMatch(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e, _err := s.CompareAndSwap("/foo", "", 100, "baz", TTLOptionSet{ExpireTime: Permanent})
	err := _err.(*etcdErr.Error)
	testutil.AssertEqual(t, err.ErrorCode, etcdErr.EcodeTestFailed)
	testutil.AssertEqual(t, err.Message, "Compare failed")
	testutil.AssertNil(t, e)
	e, _ = s.Get("/foo", false, false)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, *e.Node.Value, "bar")
}

// Ensure that the store can watch for key creation.
func TestStoreWatchCreate(t *testing.T) {
	s := newStore()
	var eidx uint64 = 0
	w, _ := s.Watch("/foo", false, false, 0)
	c := w.EventChan()
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	eidx = 1
	e := nbselect(c)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	e = nbselect(c)
	testutil.AssertNil(t, e)
}

// Ensure that the store can watch for recursive key creation.
func TestStoreWatchRecursiveCreate(t *testing.T) {
	s := newStore()
	var eidx uint64 = 0
	w, _ := s.Watch("/foo", true, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	eidx = 1
	s.Create("/foo/bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/foo/bar")
}

// Ensure that the store can watch for key updates.
func TestStoreWatchUpdate(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", false, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	eidx = 2
	s.Update("/foo", "baz", TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "update")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
}

// Ensure that the store can watch for recursive key updates.
func TestStoreWatchRecursiveUpdate(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo/bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", true, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	eidx = 2
	s.Update("/foo/bar", "baz", TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "update")
	testutil.AssertEqual(t, e.Node.Key, "/foo/bar")
}

// Ensure that the store can watch for key deletions.
func TestStoreWatchDelete(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", false, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	eidx = 2
	s.Delete("/foo", false, false)
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "delete")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
}

// Ensure that the store can watch for recursive key deletions.
func TestStoreWatchRecursiveDelete(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo/bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", true, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	eidx = 2
	s.Delete("/foo/bar", false, false)
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "delete")
	testutil.AssertEqual(t, e.Node.Key, "/foo/bar")
}

// Ensure that the store can watch for CAS updates.
func TestStoreWatchCompareAndSwap(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", false, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	eidx = 2
	s.CompareAndSwap("/foo", "bar", 0, "baz", TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "compareAndSwap")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
}

// Ensure that the store can watch for recursive CAS updates.
func TestStoreWatchRecursiveCompareAndSwap(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	s.Create("/foo/bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", true, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	eidx = 2
	s.CompareAndSwap("/foo/bar", "baz", 0, "bat", TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "compareAndSwap")
	testutil.AssertEqual(t, e.Node.Key, "/foo/bar")
}

// Ensure that the store can watch for key expiration.
func TestStoreWatchExpire(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	var eidx uint64 = 3
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: fc.Now().Add(400 * time.Millisecond)})
	s.Create("/foofoo", false, "barbarbar", false, TTLOptionSet{ExpireTime: fc.Now().Add(450 * time.Millisecond)})
	s.Create("/foodir", true, "", false, TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond)})

	w, _ := s.Watch("/", true, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	c := w.EventChan()
	e := nbselect(c)
	testutil.AssertNil(t, e)
	fc.Advance(600 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	eidx = 4
	e = nbselect(c)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "expire")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	w, _ = s.Watch("/", true, false, 5)
	eidx = 6
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	e = nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "expire")
	testutil.AssertEqual(t, e.Node.Key, "/foofoo")
	w, _ = s.Watch("/", true, false, 6)
	e = nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "expire")
	testutil.AssertEqual(t, e.Node.Key, "/foodir")
	testutil.AssertEqual(t, e.Node.Dir, true)
}

// Ensure that the store can watch for key expiration when refreshing.
func TestStoreWatchExpireRefresh(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	var eidx uint64 = 2
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	s.Create("/foofoo", false, "barbarbar", false, TTLOptionSet{ExpireTime: fc.Now().Add(1200 * time.Millisecond), Refresh: true})

	// Make sure we set watch updates when Refresh is true for newly created keys
	w, _ := s.Watch("/", true, false, 0)
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	c := w.EventChan()
	e := nbselect(c)
	testutil.AssertNil(t, e)
	fc.Advance(600 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	eidx = 3
	e = nbselect(c)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "expire")
	testutil.AssertEqual(t, e.Node.Key, "/foo")

	s.Update("/foofoo", "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	w, _ = s.Watch("/", true, false, 4)
	fc.Advance(700 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	eidx = 5 // We should skip 4 because a TTL update should occur with no watch notification if set `TTLOptionSet.Refresh` to true
	testutil.AssertEqual(t, w.StartIndex(), eidx-1)
	e = nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "expire")
	testutil.AssertEqual(t, e.Node.Key, "/foofoo")
}

// Ensure that the store can watch for key expiration when refreshing with an empty value.
func TestStoreWatchExpireEmptyRefresh(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	// Should be no-op
	fc.Advance(200 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())

	s.Update("/foo", "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	w, _ := s.Watch("/", true, false, 2)
	fc.Advance(700 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	eidx = 3 // We should skip 2 because a TTL update should occur with no watch notification if set `TTLOptionSet.Refresh` to true
	testutil.AssertEqual(t, w.StartIndex(), eidx-1)
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "expire")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
}

// Update TTL of a key (set TTLOptionSet.Refresh to false) and send notification
func TestStoreWatchNoRefresh(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	var eidx uint64 = 1
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	// Should be no-op
	fc.Advance(200 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())

	// Update key's TTL with setting `TTLOptionSet.Refresh` to false will cause an update event
	s.Update("/foo", "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: false})
	w, _ := s.Watch("/", true, false, 2)
	fc.Advance(700 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	eidx = 2
	testutil.AssertEqual(t, w.StartIndex(), eidx)
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "update")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertEqual(t, *e.PrevNode.Value, "bar")
}

// Ensure that the store can update the TTL on a value with refresh.
func TestStoreRefresh(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond)})
	s.Create("/bar", true, "bar", false, TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond)})
	_, err := s.Update("/foo", "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	testutil.AssertNil(t, err)

	_, err = s.Set("/foo", false, "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	testutil.AssertNil(t, err)

	_, err = s.Update("/bar", "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	testutil.AssertNil(t, err)

	_, err = s.CompareAndSwap("/foo", "bar", 0, "", TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond), Refresh: true})
	testutil.AssertNil(t, err)
}

// Ensure that the store can watch in streaming mode.
func TestStoreWatchStream(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	w, _ := s.Watch("/foo", false, true, 0)
	// first modification
	s.Create("/foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertEqual(t, *e.Node.Value, "bar")
	e = nbselect(w.EventChan())
	testutil.AssertNil(t, e)
	// second modification
	eidx = 2
	s.Update("/foo", "baz", TTLOptionSet{ExpireTime: Permanent})
	e = nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "update")
	testutil.AssertEqual(t, e.Node.Key, "/foo")
	testutil.AssertEqual(t, *e.Node.Value, "baz")
	e = nbselect(w.EventChan())
	testutil.AssertNil(t, e)
}

// Ensure that the store can recover from a previously saved state.
func TestStoreRecover(t *testing.T) {
	s := newStore()
	var eidx uint64 = 4
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/x", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	s.Update("/foo/x", "barbar", TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/y", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	b, err := s.Save()
	testutil.AssertNil(t, err)

	s2 := newStore()
	s2.Recovery(b)

	e, err := s.Get("/foo/x", false, false)
	testutil.AssertEqual(t, e.Node.CreatedIndex, uint64(2))
	testutil.AssertEqual(t, e.Node.ModifiedIndex, uint64(3))
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, *e.Node.Value, "barbar")

	e, err = s.Get("/foo/y", false, false)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, *e.Node.Value, "baz")
}

// Ensure that the store can recover from a previously saved state that includes an expiring key.
func TestStoreRecoverWithExpiration(t *testing.T) {
	s := newStore()
	s.clock = newFakeClock()

	fc := newFakeClock()

	var eidx uint64 = 4
	s.Create("/foo", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/x", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	s.Create("/foo/y", false, "baz", false, TTLOptionSet{ExpireTime: fc.Now().Add(5 * time.Millisecond)})
	b, err := s.Save()
	testutil.AssertNil(t, err)

	time.Sleep(10 * time.Millisecond)

	s2 := newStore()
	s2.clock = fc

	s2.Recovery(b)

	fc.Advance(600 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())

	e, err := s.Get("/foo/x", false, false)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, *e.Node.Value, "bar")

	e, err = s.Get("/foo/y", false, false)
	testutil.AssertNotNil(t, err)
	testutil.AssertNil(t, e)
}

// Ensure that the store can watch for hidden keys as long as it's an exact path match.
func TestStoreWatchCreateWithHiddenKey(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	w, _ := s.Watch("/_foo", false, false, 0)
	s.Create("/_foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/_foo")
	e = nbselect(w.EventChan())
	testutil.AssertNil(t, e)
}

// Ensure that the store doesn't see hidden key creates without an exact path match in recursive mode.
func TestStoreWatchRecursiveCreateWithHiddenKey(t *testing.T) {
	s := newStore()
	w, _ := s.Watch("/foo", true, false, 0)
	s.Create("/foo/_bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertNil(t, e)
	w, _ = s.Watch("/foo", true, false, 0)
	s.Create("/foo/_baz", true, "", false, TTLOptionSet{ExpireTime: Permanent})
	e = nbselect(w.EventChan())
	testutil.AssertNil(t, e)
	s.Create("/foo/_baz/quux", false, "quux", false, TTLOptionSet{ExpireTime: Permanent})
	e = nbselect(w.EventChan())
	testutil.AssertNil(t, e)
}

// Ensure that the store doesn't see hidden key updates.
func TestStoreWatchUpdateWithHiddenKey(t *testing.T) {
	s := newStore()
	s.Create("/_foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/_foo", false, false, 0)
	s.Update("/_foo", "baz", TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.Action, "update")
	testutil.AssertEqual(t, e.Node.Key, "/_foo")
	e = nbselect(w.EventChan())
	testutil.AssertNil(t, e)
}

// Ensure that the store doesn't see hidden key updates without an exact path match in recursive mode.
func TestStoreWatchRecursiveUpdateWithHiddenKey(t *testing.T) {
	s := newStore()
	s.Create("/foo/_bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", true, false, 0)
	s.Update("/foo/_bar", "baz", TTLOptionSet{ExpireTime: Permanent})
	e := nbselect(w.EventChan())
	testutil.AssertNil(t, e)
}

// Ensure that the store can watch for key deletions.
func TestStoreWatchDeleteWithHiddenKey(t *testing.T) {
	s := newStore()
	var eidx uint64 = 2
	s.Create("/_foo", false, "bar", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/_foo", false, false, 0)
	s.Delete("/_foo", false, false)
	e := nbselect(w.EventChan())
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "delete")
	testutil.AssertEqual(t, e.Node.Key, "/_foo")
	e = nbselect(w.EventChan())
	testutil.AssertNil(t, e)
}

// Ensure that the store doesn't see hidden key deletes without an exact path match in recursive mode.
func TestStoreWatchRecursiveDeleteWithHiddenKey(t *testing.T) {
	s := newStore()
	s.Create("/foo/_bar", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})
	w, _ := s.Watch("/foo", true, false, 0)
	s.Delete("/foo/_bar", false, false)
	e := nbselect(w.EventChan())
	testutil.AssertNil(t, e)
}

// Ensure that the store doesn't see expirations of hidden keys.
func TestStoreWatchExpireWithHiddenKey(t *testing.T) {
	s := newStore()
	fc := newFakeClock()
	s.clock = fc

	s.Create("/_foo", false, "bar", false, TTLOptionSet{ExpireTime: fc.Now().Add(500 * time.Millisecond)})
	s.Create("/foofoo", false, "barbarbar", false, TTLOptionSet{ExpireTime: fc.Now().Add(1000 * time.Millisecond)})

	w, _ := s.Watch("/", true, false, 0)
	c := w.EventChan()
	e := nbselect(c)
	testutil.AssertNil(t, e)
	fc.Advance(600 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	e = nbselect(c)
	testutil.AssertNil(t, e)
	fc.Advance(600 * time.Millisecond)
	s.DeleteExpiredKeys(fc.Now())
	e = nbselect(c)
	testutil.AssertEqual(t, e.Action, "expire")
	testutil.AssertEqual(t, e.Node.Key, "/foofoo")
}

// Ensure that the store does see hidden key creates if watching deeper than a hidden key in recursive mode.
func TestStoreWatchRecursiveCreateDeeperThanHiddenKey(t *testing.T) {
	s := newStore()
	var eidx uint64 = 1
	w, _ := s.Watch("/_foo/bar", true, false, 0)
	s.Create("/_foo/bar/baz", false, "baz", false, TTLOptionSet{ExpireTime: Permanent})

	e := nbselect(w.EventChan())
	testutil.AssertNotNil(t, e)
	testutil.AssertEqual(t, e.EtcdIndex, eidx)
	testutil.AssertEqual(t, e.Action, "create")
	testutil.AssertEqual(t, e.Node.Key, "/_foo/bar/baz")
}

// Ensure that slow consumers are handled properly.
//
// Since Watcher.EventChan() has a buffer of size 100 we can only queue 100
// event per watcher. If the consumer cannot consume the event on time and
// another event arrives, the channel is closed and event is discarded.
// This test ensures that after closing the channel, the store can continue
// to operate correctly.
func TestStoreWatchSlowConsumer(t *testing.T) {
	s := newStore()
	s.Watch("/foo", true, true, 0) // stream must be true
	// Fill watch channel with 100 events
	for i := 1; i <= 100; i++ {
		s.Set("/foo", false, string(i), TTLOptionSet{ExpireTime: Permanent}) // ok
	}
	testutil.AssertEqual(t, s.WatcherHub.count, int64(1))
	s.Set("/foo", false, "101", TTLOptionSet{ExpireTime: Permanent}) // ok
	// remove watcher
	testutil.AssertEqual(t, s.WatcherHub.count, int64(0))
	s.Set("/foo", false, "102", TTLOptionSet{ExpireTime: Permanent}) // must not panic
}

// Performs a non-blocking select on an event channel.
func nbselect(c <-chan *Event) *Event {
	select {
	case e := <-c:
		return e
	default:
		return nil
	}
}

// newFakeClock creates a new FakeClock that has been advanced to at least minExpireTime
func newFakeClock() clockwork.FakeClock {
	fc := clockwork.NewFakeClock()
	for minExpireTime.After(fc.Now()) {
		fc.Advance((0x1 << 62) * time.Nanosecond)
	}
	return fc
}
