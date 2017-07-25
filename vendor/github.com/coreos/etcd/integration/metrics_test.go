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
	"strconv"
	"testing"
	"time"

	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/pkg/testutil"
)

// TestMetricDbSizeBoot checks that the db size metric is set on boot.
func TestMetricDbSizeBoot(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	v, err := clus.Members[0].Metric("etcd_debugging_mvcc_db_total_size_in_bytes")
	if err != nil {
		t.Fatal(err)
	}

	if v == "0" {
		t.Fatalf("expected non-zero, got %q", v)
	}
}

// TestMetricDbSizeDefrag checks that the db size metric is set after defrag.
func TestMetricDbSizeDefrag(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	kvc := toGRPC(clus.Client(0)).KV
	mc := toGRPC(clus.Client(0)).Maintenance

	// expand the db size
	numPuts := 10
	putreq := &pb.PutRequest{Key: []byte("k"), Value: make([]byte, 4096)}
	for i := 0; i < numPuts; i++ {
		if _, err := kvc.Put(context.TODO(), putreq); err != nil {
			t.Fatal(err)
		}
	}

	// wait for backend txn sync
	time.Sleep(500 * time.Millisecond)

	beforeDefrag, err := clus.Members[0].Metric("etcd_debugging_mvcc_db_total_size_in_bytes")
	if err != nil {
		t.Fatal(err)
	}
	bv, err := strconv.Atoi(beforeDefrag)
	if err != nil {
		t.Fatal(err)
	}
	if expected := numPuts * len(putreq.Value); bv < expected {
		t.Fatalf("expected db size greater than %d, got %d", expected, bv)
	}

	// clear out historical keys
	creq := &pb.CompactionRequest{Revision: int64(numPuts), Physical: true}
	if _, err := kvc.Compact(context.TODO(), creq); err != nil {
		t.Fatal(err)
	}

	// defrag should give freed space back to fs
	mc.Defragment(context.TODO(), &pb.DefragmentRequest{})
	afterDefrag, err := clus.Members[0].Metric("etcd_debugging_mvcc_db_total_size_in_bytes")
	if err != nil {
		t.Fatal(err)
	}

	av, err := strconv.Atoi(afterDefrag)
	if err != nil {
		t.Fatal(err)
	}

	if bv <= av {
		t.Fatalf("expected less than %d, got %d after defrag", bv, av)
	}
}
