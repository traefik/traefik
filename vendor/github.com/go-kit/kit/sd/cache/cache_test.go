package cache

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func TestCache(t *testing.T) {
	var (
		ca    = make(closer)
		cb    = make(closer)
		c     = map[string]io.Closer{"a": ca, "b": cb}
		f     = func(instance string) (endpoint.Endpoint, io.Closer, error) { return endpoint.Nop, c[instance], nil }
		cache = New(f, log.NewNopLogger())
	)

	// Populate
	cache.Update([]string{"a", "b"})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Errorf("endpoint b closed, not good")
	case <-time.After(time.Millisecond):
		t.Logf("no closures yet, good")
	}
	if want, have := 2, len(cache.Endpoints()); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Duplicate, should be no-op
	cache.Update([]string{"a", "b"})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Errorf("endpoint b closed, not good")
	case <-time.After(time.Millisecond):
		t.Logf("no closures yet, good")
	}
	if want, have := 2, len(cache.Endpoints()); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Delete b
	go cache.Update([]string{"a"})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Logf("endpoint b closed, good")
	case <-time.After(time.Second):
		t.Errorf("didn't close the deleted instance in time")
	}
	if want, have := 1, len(cache.Endpoints()); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Delete a
	go cache.Update([]string{})
	select {
	// case <-cb: will succeed, as it's closed
	case <-ca:
		t.Logf("endpoint a closed, good")
	case <-time.After(time.Second):
		t.Errorf("didn't close the deleted instance in time")
	}
	if want, have := 0, len(cache.Endpoints()); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadFactory(t *testing.T) {
	cache := New(func(string) (endpoint.Endpoint, io.Closer, error) {
		return nil, nil, errors.New("bad factory")
	}, log.NewNopLogger())

	cache.Update([]string{"foo:1234", "bar:5678"})
	if want, have := 0, len(cache.Endpoints()); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type closer chan struct{}

func (c closer) Close() error { close(c); return nil }
