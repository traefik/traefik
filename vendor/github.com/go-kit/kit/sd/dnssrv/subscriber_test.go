package dnssrv

import (
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func TestRefresh(t *testing.T) {
	name := "some.service.internal"

	ticker := time.NewTicker(time.Second)
	ticker.Stop()
	tickc := make(chan time.Time)
	ticker.C = tickc

	var lookups uint64
	records := []*net.SRV{}
	lookup := func(service, proto, name string) (string, []*net.SRV, error) {
		t.Logf("lookup(%q, %q, %q)", service, proto, name)
		atomic.AddUint64(&lookups, 1)
		return "cname", records, nil
	}

	var generates uint64
	factory := func(instance string) (endpoint.Endpoint, io.Closer, error) {
		t.Logf("factory(%q)", instance)
		atomic.AddUint64(&generates, 1)
		return endpoint.Nop, nopCloser{}, nil
	}

	subscriber := NewSubscriberDetailed(name, ticker, lookup, factory, log.NewNopLogger())
	defer subscriber.Stop()

	// First lookup, empty
	endpoints, err := subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := uint64(1), atomic.LoadUint64(&lookups); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := uint64(0), atomic.LoadUint64(&generates); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Load some records and lookup again
	records = []*net.SRV{
		{Target: "1.0.0.1", Port: 1001},
		{Target: "1.0.0.2", Port: 1002},
		{Target: "1.0.0.3", Port: 1003},
	}
	tickc <- time.Now()

	// There is a race condition where the subscriber.Endpoints call below
	// invokes the cache before it is updated by the tick above.
	// TODO(pb): solve by running the read through the loop goroutine.
	time.Sleep(100 * time.Millisecond)

	endpoints, err = subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 3, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := uint64(2), atomic.LoadUint64(&lookups); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := uint64(len(records)), atomic.LoadUint64(&generates); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }
