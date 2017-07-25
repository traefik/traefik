// Copyright 2016 The etcd Authors
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

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"golang.org/x/net/context"
)

// TestElectionWait tests if followers can correctly wait for elections.
func TestElectionWait(t *testing.T) {
	clus := NewClusterV3(t, &ClusterConfig{Size: 3})
	defer clus.Terminate(t)

	leaders := 3
	followers := 3
	var clients []*clientv3.Client
	newClient := makeMultiNodeClients(t, clus.cluster, &clients)

	electedc := make(chan string)
	nextc := []chan struct{}{}

	// wait for all elections
	donec := make(chan struct{})
	for i := 0; i < followers; i++ {
		nextc = append(nextc, make(chan struct{}))
		go func(ch chan struct{}) {
			for j := 0; j < leaders; j++ {
				session, err := concurrency.NewSession(newClient())
				if err != nil {
					t.Error(err)
				}
				b := concurrency.NewElection(session, "test-election")

				cctx, cancel := context.WithCancel(context.TODO())
				defer cancel()
				s, ok := <-b.Observe(cctx)
				if !ok {
					t.Fatalf("could not observe election; channel closed")
				}
				electedc <- string(s.Kvs[0].Value)
				// wait for next election round
				<-ch
				session.Orphan()
			}
			donec <- struct{}{}
		}(nextc[i])
	}

	// elect some leaders
	for i := 0; i < leaders; i++ {
		go func() {
			session, err := concurrency.NewSession(newClient())
			if err != nil {
				t.Error(err)
			}
			defer session.Orphan()

			e := concurrency.NewElection(session, "test-election")
			ev := fmt.Sprintf("electval-%v", time.Now().UnixNano())
			if err := e.Campaign(context.TODO(), ev); err != nil {
				t.Fatalf("failed volunteer (%v)", err)
			}
			// wait for followers to accept leadership
			for j := 0; j < followers; j++ {
				s := <-electedc
				if s != ev {
					t.Errorf("wrong election value got %s, wanted %s", s, ev)
				}
			}
			// let next leader take over
			if err := e.Resign(context.TODO()); err != nil {
				t.Fatalf("failed resign (%v)", err)
			}
			// tell followers to start listening for next leader
			for j := 0; j < followers; j++ {
				nextc[j] <- struct{}{}
			}
		}()
	}

	// wait on followers
	for i := 0; i < followers; i++ {
		<-donec
	}

	closeClients(t, clients)
}

// TestElectionFailover tests that an election will
func TestElectionFailover(t *testing.T) {
	clus := NewClusterV3(t, &ClusterConfig{Size: 3})
	defer clus.Terminate(t)

	cctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	ss := make([]*concurrency.Session, 3)

	for i := 0; i < 3; i++ {
		var err error
		ss[i], err = concurrency.NewSession(clus.clients[i])
		if err != nil {
			t.Error(err)
		}
		defer ss[i].Orphan()
	}

	// first leader (elected)
	e := concurrency.NewElection(ss[0], "test-election")
	if err := e.Campaign(context.TODO(), "foo"); err != nil {
		t.Fatalf("failed volunteer (%v)", err)
	}

	// check first leader
	resp, ok := <-e.Observe(cctx)
	if !ok {
		t.Fatalf("could not wait for first election; channel closed")
	}
	s := string(resp.Kvs[0].Value)
	if s != "foo" {
		t.Fatalf("wrong election result. got %s, wanted foo", s)
	}

	// next leader
	electedc := make(chan struct{})
	go func() {
		ee := concurrency.NewElection(ss[1], "test-election")
		if eer := ee.Campaign(context.TODO(), "bar"); eer != nil {
			t.Fatal(eer)
		}
		electedc <- struct{}{}
	}()

	// invoke leader failover
	if err := ss[0].Close(); err != nil {
		t.Fatal(err)
	}

	// check new leader
	e = concurrency.NewElection(ss[2], "test-election")
	resp, ok = <-e.Observe(cctx)
	if !ok {
		t.Fatalf("could not wait for second election; channel closed")
	}
	s = string(resp.Kvs[0].Value)
	if s != "bar" {
		t.Fatalf("wrong election result. got %s, wanted bar", s)
	}

	// leader must ack election (otherwise, Campaign may see closed conn)
	<-electedc
}

// TestElectionSessionRelock ensures that campaigning twice on the same election
// with the same lock will Proclaim instead of deadlocking.
func TestElectionSessionRecampaign(t *testing.T) {
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)
	cli := clus.RandClient()

	session, err := concurrency.NewSession(cli)
	if err != nil {
		t.Error(err)
	}
	defer session.Orphan()

	e := concurrency.NewElection(session, "test-elect")
	if err := e.Campaign(context.TODO(), "abc"); err != nil {
		t.Fatal(err)
	}
	e2 := concurrency.NewElection(session, "test-elect")
	if err := e2.Campaign(context.TODO(), "def"); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	if resp := <-e.Observe(ctx); len(resp.Kvs) == 0 || string(resp.Kvs[0].Value) != "def" {
		t.Fatalf("expected value=%q, got response %v", "def", resp)
	}
}

// TestElectionOnPrefixOfExistingKey checks that a single
// candidate can be elected on a new key that is a prefix
// of an existing key. To wit, check for regression
// of bug #6278. https://github.com/coreos/etcd/issues/6278
//
func TestElectionOnPrefixOfExistingKey(t *testing.T) {
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	cli := clus.RandClient()
	if _, err := cli.Put(context.TODO(), "testa", "value"); err != nil {
		t.Fatal(err)
	}
	s, serr := concurrency.NewSession(cli)
	if serr != nil {
		t.Fatal(serr)
	}
	e := concurrency.NewElection(s, "test")
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	err := e.Campaign(ctx, "abc")
	cancel()
	if err != nil {
		// after 5 seconds, deadlock results in
		// 'context deadline exceeded' here.
		t.Fatal(err)
	}
}

// TestElectionOnSessionRestart tests that a quick restart of leader (resulting
// in a new session with the same lease id) does not result in loss of
// leadership.
func TestElectionOnSessionRestart(t *testing.T) {
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)
	cli := clus.RandClient()

	session, err := concurrency.NewSession(cli)
	if err != nil {
		t.Fatal(err)
	}

	e := concurrency.NewElection(session, "test-elect")
	if cerr := e.Campaign(context.TODO(), "abc"); cerr != nil {
		t.Fatal(cerr)
	}

	// ensure leader is not lost to waiter on fail-over
	waitSession, werr := concurrency.NewSession(cli)
	if werr != nil {
		t.Fatal(werr)
	}
	defer waitSession.Orphan()
	waitCtx, waitCancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer waitCancel()
	go concurrency.NewElection(waitSession, "test-elect").Campaign(waitCtx, "123")

	// simulate restart by reusing the lease from the old session
	newSession, nerr := concurrency.NewSession(cli, concurrency.WithLease(session.Lease()))
	if nerr != nil {
		t.Fatal(nerr)
	}
	defer newSession.Orphan()

	newElection := concurrency.NewElection(newSession, "test-elect")
	if ncerr := newElection.Campaign(context.TODO(), "def"); ncerr != nil {
		t.Fatal(ncerr)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if resp := <-newElection.Observe(ctx); len(resp.Kvs) == 0 || string(resp.Kvs[0].Value) != "def" {
		t.Errorf("expected value=%q, got response %v", "def", resp)
	}
}

// TestElectionObserveCompacted checks that observe can tolerate
// a leader key with a modrev less than the compaction revision.
func TestElectionObserveCompacted(t *testing.T) {
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	cli := clus.Client(0)

	session, err := concurrency.NewSession(cli)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Orphan()

	e := concurrency.NewElection(session, "test-elect")
	if cerr := e.Campaign(context.TODO(), "abc"); cerr != nil {
		t.Fatal(cerr)
	}

	presp, perr := cli.Put(context.TODO(), "foo", "bar")
	if perr != nil {
		t.Fatal(perr)
	}
	if _, cerr := cli.Compact(context.TODO(), presp.Header.Revision); cerr != nil {
		t.Fatal(cerr)
	}

	v, ok := <-e.Observe(context.TODO())
	if !ok {
		t.Fatal("failed to observe on compacted revision")
	}
	if string(v.Kvs[0].Value) != "abc" {
		t.Fatalf(`expected leader value "abc", got %q`, string(v.Kvs[0].Value))
	}
}
