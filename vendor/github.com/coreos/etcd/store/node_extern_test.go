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
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/coreos/etcd/pkg/testutil"
)

func TestNodeExternClone(t *testing.T) {
	var eNode *NodeExtern
	if g := eNode.Clone(); g != nil {
		t.Fatalf("nil.Clone=%v, want nil", g)
	}

	const (
		key string = "/foo/bar"
		ttl int64  = 123456789
		ci  uint64 = 123
		mi  uint64 = 321
	)
	var (
		val    = "some_data"
		valp   = &val
		exp    = time.Unix(12345, 67890)
		expp   = &exp
		child  = NodeExtern{}
		childp = &child
		childs = []*NodeExtern{childp}
	)

	eNode = &NodeExtern{
		Key:           key,
		TTL:           ttl,
		CreatedIndex:  ci,
		ModifiedIndex: mi,
		Value:         valp,
		Expiration:    expp,
		Nodes:         childs,
	}

	gNode := eNode.Clone()
	// Check the clone is as expected
	testutil.AssertEqual(t, gNode.Key, key)
	testutil.AssertEqual(t, gNode.TTL, ttl)
	testutil.AssertEqual(t, gNode.CreatedIndex, ci)
	testutil.AssertEqual(t, gNode.ModifiedIndex, mi)
	// values should be the same
	testutil.AssertEqual(t, *gNode.Value, val)
	testutil.AssertEqual(t, *gNode.Expiration, exp)
	testutil.AssertEqual(t, len(gNode.Nodes), len(childs))
	testutil.AssertEqual(t, *gNode.Nodes[0], child)
	// but pointers should differ
	if gNode.Value == eNode.Value {
		t.Fatalf("expected value pointers to differ, but got same!")
	}
	if gNode.Expiration == eNode.Expiration {
		t.Fatalf("expected expiration pointers to differ, but got same!")
	}
	if sameSlice(gNode.Nodes, eNode.Nodes) {
		t.Fatalf("expected nodes pointers to differ, but got same!")
	}
	// Original should be the same
	testutil.AssertEqual(t, eNode.Key, key)
	testutil.AssertEqual(t, eNode.TTL, ttl)
	testutil.AssertEqual(t, eNode.CreatedIndex, ci)
	testutil.AssertEqual(t, eNode.ModifiedIndex, mi)
	testutil.AssertEqual(t, eNode.Value, valp)
	testutil.AssertEqual(t, eNode.Expiration, expp)
	if !sameSlice(eNode.Nodes, childs) {
		t.Fatalf("expected nodes pointer to same, but got different!")
	}
	// Change the clone and ensure the original is not affected
	gNode.Key = "/baz"
	gNode.TTL = 0
	gNode.Nodes[0].Key = "uno"
	testutil.AssertEqual(t, eNode.Key, key)
	testutil.AssertEqual(t, eNode.TTL, ttl)
	testutil.AssertEqual(t, eNode.CreatedIndex, ci)
	testutil.AssertEqual(t, eNode.ModifiedIndex, mi)
	testutil.AssertEqual(t, *eNode.Nodes[0], child)
	// Change the original and ensure the clone is not affected
	eNode.Key = "/wuf"
	testutil.AssertEqual(t, eNode.Key, "/wuf")
	testutil.AssertEqual(t, gNode.Key, "/baz")
}

func sameSlice(a, b []*NodeExtern) bool {
	ah := (*reflect.SliceHeader)(unsafe.Pointer(&a))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	return *ah == *bh
}
