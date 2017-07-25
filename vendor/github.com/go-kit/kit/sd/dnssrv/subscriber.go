package dnssrv

import (
	"fmt"
	"net"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/cache"
)

// Subscriber yields endpoints taken from the named DNS SRV record. The name is
// resolved on a fixed schedule. Priorities and weights are ignored.
type Subscriber struct {
	name   string
	cache  *cache.Cache
	logger log.Logger
	quit   chan struct{}
}

// NewSubscriber returns a DNS SRV subscriber.
func NewSubscriber(
	name string,
	ttl time.Duration,
	factory sd.Factory,
	logger log.Logger,
) *Subscriber {
	return NewSubscriberDetailed(name, time.NewTicker(ttl), net.LookupSRV, factory, logger)
}

// NewSubscriberDetailed is the same as NewSubscriber, but allows users to
// provide an explicit lookup refresh ticker instead of a TTL, and specify the
// lookup function instead of using net.LookupSRV.
func NewSubscriberDetailed(
	name string,
	refresh *time.Ticker,
	lookup Lookup,
	factory sd.Factory,
	logger log.Logger,
) *Subscriber {
	p := &Subscriber{
		name:   name,
		cache:  cache.New(factory, logger),
		logger: logger,
		quit:   make(chan struct{}),
	}

	instances, err := p.resolve(lookup)
	if err == nil {
		logger.Log("name", name, "instances", len(instances))
	} else {
		logger.Log("name", name, "err", err)
	}
	p.cache.Update(instances)

	go p.loop(refresh, lookup)
	return p
}

// Stop terminates the Subscriber.
func (p *Subscriber) Stop() {
	close(p.quit)
}

func (p *Subscriber) loop(t *time.Ticker, lookup Lookup) {
	defer t.Stop()
	for {
		select {
		case <-t.C:
			instances, err := p.resolve(lookup)
			if err != nil {
				p.logger.Log("name", p.name, "err", err)
				continue // don't replace potentially-good with bad
			}
			p.cache.Update(instances)

		case <-p.quit:
			return
		}
	}
}

// Endpoints implements the Subscriber interface.
func (p *Subscriber) Endpoints() ([]endpoint.Endpoint, error) {
	return p.cache.Endpoints(), nil
}

func (p *Subscriber) resolve(lookup Lookup) ([]string, error) {
	_, addrs, err := lookup("", "", p.name)
	if err != nil {
		return []string{}, err
	}
	instances := make([]string, len(addrs))
	for i, addr := range addrs {
		instances[i] = net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))
	}
	return instances, nil
}
