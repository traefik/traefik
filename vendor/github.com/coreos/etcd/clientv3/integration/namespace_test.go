// Copyright 2017 The etcd Authors
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

package integration

import (
	"context"
	"reflect"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/coreos/etcd/integration"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/coreos/etcd/pkg/testutil"
)

func TestNamespacePutGet(t *testing.T) {
	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	c := clus.Client(0)
	nsKV := namespace.NewKV(c.KV, "foo/")

	if _, err := nsKV.Put(context.TODO(), "abc", "bar"); err != nil {
		t.Fatal(err)
	}
	resp, err := nsKV.Get(context.TODO(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if string(resp.Kvs[0].Key) != "abc" {
		t.Errorf("expected key=%q, got key=%q", "abc", resp.Kvs[0].Key)
	}

	resp, err = c.Get(context.TODO(), "foo/abc")
	if err != nil {
		t.Fatal(err)
	}
	if string(resp.Kvs[0].Value) != "bar" {
		t.Errorf("expected value=%q, got value=%q", "bar", resp.Kvs[0].Value)
	}
}

func TestNamespaceWatch(t *testing.T) {
	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	c := clus.Client(0)
	nsKV := namespace.NewKV(c.KV, "foo/")
	nsWatcher := namespace.NewWatcher(c.Watcher, "foo/")

	if _, err := nsKV.Put(context.TODO(), "abc", "bar"); err != nil {
		t.Fatal(err)
	}

	nsWch := nsWatcher.Watch(context.TODO(), "abc", clientv3.WithRev(1))
	wkv := &mvccpb.KeyValue{Key: []byte("abc"), Value: []byte("bar"), CreateRevision: 2, ModRevision: 2, Version: 1}
	if wr := <-nsWch; len(wr.Events) != 1 || !reflect.DeepEqual(wr.Events[0].Kv, wkv) {
		t.Errorf("expected namespaced event %+v, got %+v", wkv, wr.Events[0].Kv)
	}

	wch := c.Watch(context.TODO(), "foo/abc", clientv3.WithRev(1))
	wkv = &mvccpb.KeyValue{Key: []byte("foo/abc"), Value: []byte("bar"), CreateRevision: 2, ModRevision: 2, Version: 1}
	if wr := <-wch; len(wr.Events) != 1 || !reflect.DeepEqual(wr.Events[0].Kv, wkv) {
		t.Errorf("expected unnamespaced event %+v, got %+v", wkv, wr)
	}

	// let client close teardown namespace watch
	c.Watcher = nsWatcher
}
