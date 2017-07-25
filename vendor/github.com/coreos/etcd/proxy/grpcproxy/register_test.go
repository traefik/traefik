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

package grpcproxy

import (
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/naming"
	"github.com/coreos/etcd/integration"
	"github.com/coreos/etcd/pkg/testutil"

	gnaming "google.golang.org/grpc/naming"
)

func TestRegister(t *testing.T) {
	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)
	cli := clus.Client(0)
	paddr := clus.Members[0].GRPCAddr()

	testPrefix := "test-name"
	wa := createWatcher(t, cli, testPrefix)
	ups, err := wa.Next()
	if err != nil {
		t.Fatal(err)
	}
	if len(ups) != 0 {
		t.Fatalf("len(ups) expected 0, got %d (%v)", len(ups), ups)
	}

	donec := Register(cli, testPrefix, paddr, 5)

	ups, err = wa.Next()
	if err != nil {
		t.Fatal(err)
	}
	if len(ups) != 1 {
		t.Fatalf("len(ups) expected 1, got %d (%v)", len(ups), ups)
	}
	if ups[0].Addr != paddr {
		t.Fatalf("ups[0].Addr expected %q, got %q", paddr, ups[0].Addr)
	}

	cli.Close()
	clus.TakeClient(0)
	select {
	case <-donec:
	case <-time.After(5 * time.Second):
		t.Fatal("donec 'register' did not return in time")
	}
}

func createWatcher(t *testing.T, c *clientv3.Client, prefix string) gnaming.Watcher {
	gr := &naming.GRPCResolver{Client: c}
	watcher, err := gr.Resolve(prefix)
	if err != nil {
		t.Fatalf("failed to resolve %q (%v)", prefix, err)
	}
	return watcher
}
