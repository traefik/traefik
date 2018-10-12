package safe

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cenk/backoff"
)

func TestNewPoolContext(t *testing.T) {
	type testKeyType string
	testKey := testKeyType("test")

	ctx := context.WithValue(context.Background(), testKey, "test")
	p := NewPool(ctx)

	retCtx := p.Ctx()

	retCtxVal, ok := retCtx.Value(testKey).(string)
	if !ok || retCtxVal != "test" {
		t.Errorf("Pool.Ctx() did not return a derived context, got %#v, expected context with test value", retCtx)
	}
}

type fakeRoutine struct {
	sync.Mutex
	started  bool
	startSig chan bool
}

func newFakeRoutine() *fakeRoutine {
	return &fakeRoutine{
		startSig: make(chan bool),
	}
}

func (tr *fakeRoutine) routineCtx(ctx context.Context) {
	tr.Lock()
	tr.started = true
	tr.Unlock()
	tr.startSig <- true
	<-ctx.Done()
}

func (tr *fakeRoutine) routine(stop chan bool) {
	tr.Lock()
	tr.started = true
	tr.Unlock()
	tr.startSig <- true
	<-stop
}

func TestPoolWithCtx(t *testing.T) {
	testRoutine := newFakeRoutine()

	testCases := []struct {
		desc string
		fn   func(*Pool)
	}{
		{
			desc: "GoCtx()",
			fn: func(p *Pool) {
				p.GoCtx(testRoutine.routineCtx)
			},
		},
		{
			desc: "AddGoCtx()",
			fn: func(p *Pool) {
				p.AddGoCtx(testRoutine.routineCtx)
				p.Start()
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			// These subtests cannot be run in parallel, since the testRoutine
			// is shared across the subtests.
			p := NewPool(context.Background())
			timer := time.NewTimer(500 * time.Millisecond)
			defer timer.Stop()

			test.fn(p)
			defer p.Cleanup()
			if len(p.routinesCtx) != 1 {
				t.Fatalf("After %s, Pool did have %d goroutineCtxs, expected 1", test.desc, len(p.routinesCtx))
			}

			testDone := make(chan bool, 1)
			go func() {
				<-testRoutine.startSig
				p.Cleanup()
				testDone <- true
			}()

			select {
			case <-timer.C:
				testRoutine.Lock()
				defer testRoutine.Unlock()
				t.Fatalf("Pool test did not complete in time, goroutine started equals '%t'", testRoutine.started)
			case <-testDone:
				return
			}
		})
	}
}

func TestPoolWithStopChan(t *testing.T) {
	testRoutine := newFakeRoutine()

	p := NewPool(context.Background())

	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()

	p.Go(testRoutine.routine)
	if len(p.routines) != 1 {
		t.Fatalf("After Pool.Go(func), Pool did have %d goroutines, expected 1", len(p.routines))
	}

	testDone := make(chan bool, 1)
	go func() {
		<-testRoutine.startSig
		p.Cleanup()
		testDone <- true
	}()

	select {
	case <-timer.C:
		testRoutine.Lock()
		defer testRoutine.Unlock()
		t.Fatalf("Pool test did not complete in time, goroutine started equals '%t'", testRoutine.started)
	case <-testDone:
		return
	}
}

func TestPoolStartWithStopChan(t *testing.T) {
	testRoutine := newFakeRoutine()

	p := NewPool(context.Background())

	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()

	// Insert the stopped test goroutine via private fields into the Pool.
	// There currently is no way to insert a routine via exported funcs that is not started immediately.
	p.lock.Lock()
	newRoutine := routine{
		goroutine: testRoutine.routine,
	}
	p.routines = append(p.routines, newRoutine)
	p.lock.Unlock()
	p.Start()

	testDone := make(chan bool, 1)
	go func() {
		<-testRoutine.startSig
		p.Cleanup()
		testDone <- true
	}()
	select {
	case <-timer.C:
		testRoutine.Lock()
		defer testRoutine.Unlock()
		t.Fatalf("Pool.Start() did not complete in time, goroutine started equals '%t'", testRoutine.started)
	case <-testDone:
		return
	}
}

func TestGoroutineRecover(t *testing.T) {
	// if recover fails the test will panic
	Go(func() {
		panic("BOOM")
	})
}

func TestOperationWithRecover(t *testing.T) {
	operation := func() error {
		return nil
	}
	err := backoff.Retry(OperationWithRecover(operation), &backoff.StopBackOff{})
	if err != nil {
		t.Fatalf("Error in OperationWithRecover: %s", err)
	}
}

func TestOperationWithRecoverPanic(t *testing.T) {
	operation := func() error {
		panic("BOOM")
	}
	err := backoff.Retry(OperationWithRecover(operation), &backoff.StopBackOff{})
	if err == nil {
		t.Fatalf("Error in OperationWithRecover: %s", err)
	}
}

func TestOperationWithRecoverError(t *testing.T) {
	operation := func() error {
		return fmt.Errorf("ERROR")
	}
	err := backoff.Retry(OperationWithRecover(operation), &backoff.StopBackOff{})
	if err == nil {
		t.Fatalf("Error in OperationWithRecover: %s", err)
	}
}
