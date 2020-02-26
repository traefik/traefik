package safe

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
)

func TestNewPoolContext(t *testing.T) {
	type testKeyType string

	testKey := testKeyType("test")

	ctx := context.WithValue(context.Background(), testKey, "test")
	p := NewPool(ctx)

	p.GoCtx(func(ctx context.Context) {
		retCtxVal, ok := ctx.Value(testKey).(string)
		if !ok || retCtxVal != "test" {
			t.Errorf("Pool.Ctx() did not return a derived context, got %#v, expected context with test value", ctx)
		}
	})
	p.Stop()
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
			defer p.Stop()

			testDone := make(chan bool, 1)
			go func() {
				<-testRoutine.startSig
				p.Stop()
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

func TestPoolCleanupWithGoPanicking(t *testing.T) {
	p := NewPool(context.Background())

	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()

	p.GoCtx(func(ctx context.Context) {
		panic("BOOM")
	})

	testDone := make(chan bool, 1)
	go func() {
		p.Stop()
		testDone <- true
	}()

	select {
	case <-timer.C:
		t.Fatalf("Pool.Cleanup() did not complete in time with a panicking goroutine")
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
