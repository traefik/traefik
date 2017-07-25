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

package clientv3

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/pkg/testutil"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	endpoints = []string{"localhost:2379", "localhost:22379", "localhost:32379"}
)

func TestBalancerGetUnblocking(t *testing.T) {
	sb := newSimpleBalancer(endpoints)
	defer sb.Close()
	if addrs := <-sb.Notify(); len(addrs) != len(endpoints) {
		t.Errorf("Initialize newSimpleBalancer should have triggered Notify() chan, but it didn't")
	}
	unblockingOpts := grpc.BalancerGetOptions{BlockingWait: false}

	_, _, err := sb.Get(context.Background(), unblockingOpts)
	if err != ErrNoAddrAvilable {
		t.Errorf("Get() with no up endpoints should return ErrNoAddrAvailable, got: %v", err)
	}

	down1 := sb.Up(grpc.Address{Addr: endpoints[1]})
	if addrs := <-sb.Notify(); len(addrs) != 1 {
		t.Errorf("first Up() should have triggered balancer to send the first connected address via Notify chan so that other connections can be closed")
	}
	down2 := sb.Up(grpc.Address{Addr: endpoints[2]})
	addrFirst, putFun, err := sb.Get(context.Background(), unblockingOpts)
	if err != nil {
		t.Errorf("Get() with up endpoints should success, got %v", err)
	}
	if addrFirst.Addr != endpoints[1] {
		t.Errorf("Get() didn't return expected address, got %v", addrFirst)
	}
	if putFun == nil {
		t.Errorf("Get() returned unexpected nil put function")
	}
	addrSecond, _, _ := sb.Get(context.Background(), unblockingOpts)
	if addrFirst.Addr != addrSecond.Addr {
		t.Errorf("Get() didn't return the same address as previous call, got %v and %v", addrFirst, addrSecond)
	}

	down1(errors.New("error"))
	if addrs := <-sb.Notify(); len(addrs) != len(endpoints) {
		t.Errorf("closing the only connection should triggered balancer to send the all endpoints via Notify chan so that we can establish a connection")
	}
	down2(errors.New("error"))
	_, _, err = sb.Get(context.Background(), unblockingOpts)
	if err != ErrNoAddrAvilable {
		t.Errorf("Get() with no up endpoints should return ErrNoAddrAvailable, got: %v", err)
	}
}

func TestBalancerGetBlocking(t *testing.T) {
	sb := newSimpleBalancer(endpoints)
	defer sb.Close()
	if addrs := <-sb.Notify(); len(addrs) != len(endpoints) {
		t.Errorf("Initialize newSimpleBalancer should have triggered Notify() chan, but it didn't")
	}
	blockingOpts := grpc.BalancerGetOptions{BlockingWait: true}

	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*100)
	_, _, err := sb.Get(ctx, blockingOpts)
	if err != context.DeadlineExceeded {
		t.Errorf("Get() with no up endpoints should timeout, got %v", err)
	}

	downC := make(chan func(error), 1)

	go func() {
		// ensure sb.Up() will be called after sb.Get() to see if Up() releases blocking Get()
		time.Sleep(time.Millisecond * 100)
		f := sb.Up(grpc.Address{Addr: endpoints[1]})
		if addrs := <-sb.Notify(); len(addrs) != 1 {
			t.Errorf("first Up() should have triggered balancer to send the first connected address via Notify chan so that other connections can be closed")
		}
		downC <- f
	}()
	addrFirst, putFun, err := sb.Get(context.Background(), blockingOpts)
	if err != nil {
		t.Errorf("Get() with up endpoints should success, got %v", err)
	}
	if addrFirst.Addr != endpoints[1] {
		t.Errorf("Get() didn't return expected address, got %v", addrFirst)
	}
	if putFun == nil {
		t.Errorf("Get() returned unexpected nil put function")
	}
	down1 := <-downC

	down2 := sb.Up(grpc.Address{Addr: endpoints[2]})
	addrSecond, _, _ := sb.Get(context.Background(), blockingOpts)
	if addrFirst.Addr != addrSecond.Addr {
		t.Errorf("Get() didn't return the same address as previous call, got %v and %v", addrFirst, addrSecond)
	}

	down1(errors.New("error"))
	if addrs := <-sb.Notify(); len(addrs) != len(endpoints) {
		t.Errorf("closing the only connection should triggered balancer to send the all endpoints via Notify chan so that we can establish a connection")
	}
	down2(errors.New("error"))
	ctx, _ = context.WithTimeout(context.Background(), time.Millisecond*100)
	_, _, err = sb.Get(ctx, blockingOpts)
	if err != context.DeadlineExceeded {
		t.Errorf("Get() with no up endpoints should timeout, got %v", err)
	}
}

// TestBalancerDoNotBlockOnClose ensures that balancer and grpc don't deadlock each other
// due to rapid open/close conn. The deadlock causes balancer.Close() to block forever.
// See issue: https://github.com/coreos/etcd/issues/7283 for more detail.
func TestBalancerDoNotBlockOnClose(t *testing.T) {
	defer testutil.AfterTest(t)

	kcl := newKillConnListener(t, 3)
	defer kcl.close()

	for i := 0; i < 5; i++ {
		sb := newSimpleBalancer(kcl.endpoints())
		conn, err := grpc.Dial("", grpc.WithInsecure(), grpc.WithBalancer(sb))
		if err != nil {
			t.Fatal(err)
		}
		kvc := pb.NewKVClient(conn)
		<-sb.readyc

		var wg sync.WaitGroup
		wg.Add(100)
		cctx, cancel := context.WithCancel(context.TODO())
		for j := 0; j < 100; j++ {
			go func() {
				defer wg.Done()
				kvc.Range(cctx, &pb.RangeRequest{}, grpc.FailFast(false))
			}()
		}
		// balancer.Close() might block
		// if balancer and grpc deadlock each other.
		bclosec, cclosec := make(chan struct{}), make(chan struct{})
		go func() {
			defer close(bclosec)
			sb.Close()
		}()
		go func() {
			defer close(cclosec)
			conn.Close()
		}()
		select {
		case <-bclosec:
		case <-time.After(3 * time.Second):
			testutil.FatalStack(t, "balancer close timeout")
		}
		select {
		case <-cclosec:
		case <-time.After(3 * time.Second):
			t.Fatal("grpc conn close timeout")
		}

		cancel()
		wg.Wait()
	}
}

// killConnListener listens incoming conn and kills it immediately.
type killConnListener struct {
	wg    sync.WaitGroup
	eps   []string
	stopc chan struct{}
	t     *testing.T
}

func newKillConnListener(t *testing.T, size int) *killConnListener {
	kcl := &killConnListener{stopc: make(chan struct{}), t: t}

	for i := 0; i < size; i++ {
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Fatal(err)
		}
		kcl.eps = append(kcl.eps, ln.Addr().String())
		kcl.wg.Add(1)
		go kcl.listen(ln)
	}
	return kcl
}

func (kcl *killConnListener) endpoints() []string {
	return kcl.eps
}

func (kcl *killConnListener) listen(l net.Listener) {
	go func() {
		defer kcl.wg.Done()
		for {
			conn, err := l.Accept()
			select {
			case <-kcl.stopc:
				return
			default:
			}
			if err != nil {
				kcl.t.Fatal(err)
			}
			time.Sleep(1 * time.Millisecond)
			conn.Close()
		}
	}()
	<-kcl.stopc
	l.Close()
}

func (kcl *killConnListener) close() {
	close(kcl.stopc)
	kcl.wg.Wait()
}
