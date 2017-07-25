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
	"fmt"
	"testing"
	"time"

	epb "github.com/coreos/etcd/etcdserver/api/v3election/v3electionpb"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/pkg/testutil"
	"golang.org/x/net/context"
)

// TestV3ElectionCampaign checks that Campaign will not give
// simultaneous leadership to multiple campaigners.
func TestV3ElectionCampaign(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	lease1, err1 := toGRPC(clus.RandClient()).Lease.LeaseGrant(context.TODO(), &pb.LeaseGrantRequest{TTL: 30})
	if err1 != nil {
		t.Fatal(err1)
	}
	lease2, err2 := toGRPC(clus.RandClient()).Lease.LeaseGrant(context.TODO(), &pb.LeaseGrantRequest{TTL: 30})
	if err2 != nil {
		t.Fatal(err2)
	}

	lc := toGRPC(clus.Client(0)).Election
	req1 := &epb.CampaignRequest{Name: []byte("foo"), Lease: lease1.ID, Value: []byte("abc")}
	l1, lerr1 := lc.Campaign(context.TODO(), req1)
	if lerr1 != nil {
		t.Fatal(lerr1)
	}

	campaignc := make(chan struct{})
	go func() {
		defer close(campaignc)
		req2 := &epb.CampaignRequest{Name: []byte("foo"), Lease: lease2.ID, Value: []byte("def")}
		l2, lerr2 := lc.Campaign(context.TODO(), req2)
		if lerr2 != nil {
			t.Fatal(lerr2)
		}
		if l1.Header.Revision >= l2.Header.Revision {
			t.Fatalf("expected l1 revision < l2 revision, got %d >= %d", l1.Header.Revision, l2.Header.Revision)
		}
	}()

	select {
	case <-time.After(200 * time.Millisecond):
	case <-campaignc:
		t.Fatalf("got leadership before resign")
	}

	if _, uerr := lc.Resign(context.TODO(), &epb.ResignRequest{Leader: l1.Leader}); uerr != nil {
		t.Fatal(uerr)
	}

	select {
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("campaigner unelected after resign")
	case <-campaignc:
	}

	lval, lverr := lc.Leader(context.TODO(), &epb.LeaderRequest{Name: []byte("foo")})
	if lverr != nil {
		t.Fatal(lverr)
	}

	if string(lval.Kv.Value) != "def" {
		t.Fatalf("got election value %q, expected %q", string(lval.Kv.Value), "def")
	}
}

// TestV3ElectionObserve checks that an Observe stream receives
// proclamations from different leaders uninterrupted.
func TestV3ElectionObserve(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	lc := toGRPC(clus.Client(0)).Election

	// observe leadership events
	observec := make(chan struct{})
	go func() {
		defer close(observec)
		s, err := lc.Observe(context.Background(), &epb.LeaderRequest{Name: []byte("foo")})
		observec <- struct{}{}
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			resp, rerr := s.Recv()
			if rerr != nil {
				t.Fatal(rerr)
			}
			respV := 0
			fmt.Sscanf(string(resp.Kv.Value), "%d", &respV)
			// leader transitions should not go backwards
			if respV < i {
				t.Fatalf(`got observe value %q, expected >= "%d"`, string(resp.Kv.Value), i)
			}
			i = respV
		}
	}()

	select {
	case <-observec:
	case <-time.After(time.Second):
		t.Fatalf("observe stream took too long to start")
	}

	lease1, err1 := toGRPC(clus.RandClient()).Lease.LeaseGrant(context.TODO(), &pb.LeaseGrantRequest{TTL: 30})
	if err1 != nil {
		t.Fatal(err1)
	}
	c1, cerr1 := lc.Campaign(context.TODO(), &epb.CampaignRequest{Name: []byte("foo"), Lease: lease1.ID, Value: []byte("0")})
	if cerr1 != nil {
		t.Fatal(cerr1)
	}

	// overlap other leader so it waits on resign
	leader2c := make(chan struct{})
	go func() {
		defer close(leader2c)

		lease2, err2 := toGRPC(clus.RandClient()).Lease.LeaseGrant(context.TODO(), &pb.LeaseGrantRequest{TTL: 30})
		if err2 != nil {
			t.Fatal(err2)
		}
		c2, cerr2 := lc.Campaign(context.TODO(), &epb.CampaignRequest{Name: []byte("foo"), Lease: lease2.ID, Value: []byte("5")})
		if cerr2 != nil {
			t.Fatal(cerr2)
		}
		for i := 6; i < 10; i++ {
			v := []byte(fmt.Sprintf("%d", i))
			req := &epb.ProclaimRequest{Leader: c2.Leader, Value: v}
			if _, err := lc.Proclaim(context.TODO(), req); err != nil {
				t.Fatal(err)
			}
		}
	}()

	for i := 1; i < 5; i++ {
		v := []byte(fmt.Sprintf("%d", i))
		req := &epb.ProclaimRequest{Leader: c1.Leader, Value: v}
		if _, err := lc.Proclaim(context.TODO(), req); err != nil {
			t.Fatal(err)
		}
	}
	// start second leader
	lc.Resign(context.TODO(), &epb.ResignRequest{Leader: c1.Leader})

	select {
	case <-observec:
	case <-time.After(time.Second):
		t.Fatalf("observe did not observe all events in time")
	}

	<-leader2c
}
