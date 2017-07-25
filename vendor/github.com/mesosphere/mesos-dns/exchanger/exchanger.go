package exchanger

import (
	"log"
	"time"

	"github.com/mesosphere/mesos-dns/logging"
	"github.com/miekg/dns"
)

// Exchanger is an interface capturing a dns.Client Exchange method.
type Exchanger interface {
	// Exchange performs an synchronous query. It sends the message m to the address
	// contained in addr (host:port) and waits for a reply.
	Exchange(m *dns.Msg, addr string) (r *dns.Msg, rtt time.Duration, err error)
}

// Func is a function type that implements the Exchanger interface.
type Func func(*dns.Msg, string) (*dns.Msg, time.Duration, error)

// Exchange implements the Exchanger interface.
func (f Func) Exchange(m *dns.Msg, addr string) (*dns.Msg, time.Duration, error) {
	return f(m, addr)
}

// A Decorator adds a layer of behaviour to a given Exchanger.
type Decorator func(Exchanger) Exchanger

// Decorate decorates an Exchanger with the given Decorators.
func Decorate(ex Exchanger, ds ...Decorator) Exchanger {
	decorated := ex
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}

// ErrorLogging returns a Decorator which logs an Exchanger's errors to the given
// logger.
func ErrorLogging(l *log.Logger) Decorator {
	return func(ex Exchanger) Exchanger {
		return Func(func(m *dns.Msg, a string) (r *dns.Msg, rtt time.Duration, err error) {
			defer func() {
				if err != nil {
					l.Printf("%v: exchanging %#v with %q", err, m, a)
				}
			}()
			return ex.Exchange(m, a)
		})
	}
}

// Instrumentation returns a Decorator which instruments an Exchanger with the given
// counters.
func Instrumentation(total, success, failure logging.Counter) Decorator {
	return func(ex Exchanger) Exchanger {
		return Func(func(m *dns.Msg, a string) (r *dns.Msg, rtt time.Duration, err error) {
			defer func() {
				if total.Inc(); err != nil {
					failure.Inc()
				} else {
					success.Inc()
				}
			}()
			return ex.Exchange(m, a)
		})
	}
}
