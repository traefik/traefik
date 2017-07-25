package exchanger

import (
	"bytes"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/mesosphere/mesos-dns/logging"
	"github.com/miekg/dns"
)

func TestErrorLogging(t *testing.T) {
	{ // with error
		var buf bytes.Buffer
		_, _, _ = ErrorLogging(log.New(&buf, "", 0))(
			stub(exchanged{err: errors.New("timeout")})).Exchange(nil, "1.2.3.4")

		want := "timeout: exchanging (*dns.Msg)(nil) with \"1.2.3.4\"\n"
		if got := buf.String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
	{ // no error
		var buf bytes.Buffer
		_, _, _ = ErrorLogging(log.New(&buf, "", 0))(
			stub(exchanged{})).Exchange(nil, "1.2.3.4")

		if got, want := buf.String(), ""; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestInstrumentation(t *testing.T) {
	{ // with error
		var total, success, failure logging.LogCounter
		_, _, _ = Instrumentation(&total, &success, &failure)(
			stub(exchanged{err: errors.New("timeout")})).Exchange(nil, "1.2.3.4")

		want := []string{"1", "0", "1"}
		for i, c := range []*logging.LogCounter{&total, &success, &failure} {
			if got, want := c.String(), want[i]; got != want {
				t.Errorf("test #%d: got %q, want %q", i, got, want)
			}
		}
	}
	{ // no error
		var total, success, failure logging.LogCounter
		_, _, _ = Instrumentation(&total, &success, &failure)(
			stub(exchanged{})).Exchange(nil, "1.2.3.4")

		want := []string{"1", "1", "0"}
		for i, c := range []*logging.LogCounter{&total, &success, &failure} {
			if got, want := c.String(), want[i]; got != want {
				t.Errorf("test #%d: got %q, want %q", i, got, want)
			}
		}
	}
}

func stubs(ed ...exchanged) []Exchanger {
	exs := make([]Exchanger, len(ed))
	for i := range ed {
		exs[i] = stub(ed[i])
	}
	return exs
}

func stub(e exchanged) Exchanger {
	return Func(func(*dns.Msg, string) (*dns.Msg, time.Duration, error) {
		return e.m, e.rtt, e.err
	})
}

type exchanged struct {
	m   *dns.Msg
	rtt time.Duration
	err error
}
