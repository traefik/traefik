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
	"net"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/integration"
	"github.com/coreos/etcd/pkg/testutil"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestClusterProxyMemberList(t *testing.T) {
	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	cts := newClusterProxyServer([]string{clus.Members[0].GRPCAddr()}, t)
	defer cts.close(t)

	cfg := clientv3.Config{
		Endpoints:   []string{cts.caddr},
		DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		t.Fatalf("err %v, want nil", err)
	}
	defer client.Close()

	// wait some time for register-loop to write keys
	time.Sleep(time.Second)

	var mresp *clientv3.MemberListResponse
	mresp, err = client.Cluster.MemberList(context.Background())
	if err != nil {
		t.Fatalf("err %v, want nil", err)
	}

	if len(mresp.Members) != 1 {
		t.Fatalf("len(mresp.Members) expected 1, got %d (%+v)", len(mresp.Members), mresp.Members)
	}
	if len(mresp.Members[0].ClientURLs) != 1 {
		t.Fatalf("len(mresp.Members[0].ClientURLs) expected 1, got %d (%+v)", len(mresp.Members[0].ClientURLs), mresp.Members[0].ClientURLs[0])
	}
	if mresp.Members[0].ClientURLs[0] != cts.caddr {
		t.Fatalf("mresp.Members[0].ClientURLs[0] expected %q, got %q", cts.caddr, mresp.Members[0].ClientURLs[0])
	}
}

type clusterproxyTestServer struct {
	cp     pb.ClusterServer
	c      *clientv3.Client
	server *grpc.Server
	l      net.Listener
	donec  <-chan struct{}
	caddr  string
}

func (cts *clusterproxyTestServer) close(t *testing.T) {
	cts.server.Stop()
	cts.l.Close()
	cts.c.Close()
	select {
	case <-cts.donec:
		return
	case <-time.After(5 * time.Second):
		t.Fatalf("register-loop took too long to return")
	}
}

func newClusterProxyServer(endpoints []string, t *testing.T) *clusterproxyTestServer {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	cts := &clusterproxyTestServer{
		c: client,
	}
	cts.l, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var opts []grpc.ServerOption
	cts.server = grpc.NewServer(opts...)
	go cts.server.Serve(cts.l)

	// wait some time for free port 0 to be resolved
	time.Sleep(500 * time.Millisecond)

	Register(client, "test-prefix", cts.l.Addr().String(), 7)
	cts.cp, cts.donec = NewClusterProxy(client, cts.l.Addr().String(), "test-prefix")
	cts.caddr = cts.l.Addr().String()
	pb.RegisterClusterServer(cts.server, cts.cp)

	return cts
}
