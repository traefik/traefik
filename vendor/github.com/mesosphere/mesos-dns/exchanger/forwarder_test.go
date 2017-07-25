package exchanger

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	. "github.com/mesosphere/mesos-dns/dnstest"
	"github.com/miekg/dns"
)

func TestForwarder(t *testing.T) {
	exs := func(e exchanged, protos ...string) map[string]Exchanger {
		es := make(map[string]Exchanger, len(protos))
		for _, proto := range protos {
			es[proto] = stub(e)
		}
		return es
	}

	msg := Message(Question("foo.bar", dns.TypeA))
	for i, tt := range []struct {
		addrs []string
		exs   map[string]Exchanger
		proto string
		r     *dns.Msg
		err   error
	}{
		{ // no matching protocol
			nil, exs(exchanged{}, "udp"), "tcp", nil, &ForwardError{nil, "tcp"},
		},
		{ // matching protocol, no addrs
			nil, exs(exchanged{}, "udp"), "udp", nil, &ForwardError{nil, "udp"},
		},
		{ // matching protocol, no addrs
			[]string{}, exs(exchanged{}, "udp"), "udp", nil, &ForwardError{[]string{}, "udp"},
		},
		{ // matching protocol, one addr, no error exchanging
			addrs: []string{"1.2.3.4"},
			exs:   exs(exchanged{m: msg}, "udp"),
			proto: "udp",
			r:     msg,
		},
		{ // matching protocol, one addr, error exchanging
			addrs: []string{"1.2.3.4"},
			exs:   exs(exchanged{err: errors.New("timeout")}, "udp"),
			proto: "udp",
			err:   errors.New("timeout"),
		},
		{ // matching protocol, two addrs, error exchanging with the first only
			addrs: []string{"1.2.3.4", "2.3.4.5"},
			exs: map[string]Exchanger{
				"udp": Func(func(_ *dns.Msg, a string) (*dns.Msg, time.Duration, error) {
					switch a {
					case "1.2.3.4":
						return nil, 0, errors.New("timeout")
					default:
						return msg, 0, nil
					}
				}),
			},
			proto: "udp",
			r:     msg,
		},
		{ // matching protocol, two addrs, error exchanging with all of them
			addrs: []string{"1.2.3.4", "2.3.4.5"},
			exs: map[string]Exchanger{
				"udp": Func(func(_ *dns.Msg, a string) (*dns.Msg, time.Duration, error) {
					switch a {
					case "1.2.3.4":
						return nil, 0, errors.New("timeout")
					default:
						return nil, 0, errors.New("eof")
					}
				}),
			},
			proto: "udp",
			err:   errors.New("eof"),
		},
	} {
		var got forwarded
		got.r, got.err = NewForwarder(tt.addrs, tt.exs).Forward(nil, tt.proto)
		if want := (forwarded{r: tt.r, err: tt.err}); !reflect.DeepEqual(got, want) {
			t.Logf("test #%d\n", i)
			t.Error(pretty.Compare(got, want))
		}
	}
}

type forwarded struct {
	r   *dns.Msg
	err error
}
