package tcp

import (
	"context"
	"errors"
	"sync"

	"github.com/rs/zerolog/log"
)

var errNoServersInPool = errors.New("no servers in the pool")

type server struct {
	Handler

	name   string
	weight int
}

// WRRLoadBalancer is a naive RoundRobin load balancer for TCP services.
type WRRLoadBalancer struct {
	// serversMu is a mutex to protect the handlers slice and the status.
	serversMu sync.Mutex
	servers   []server
	// status is a record of which child services of the Balancer are healthy, keyed
	// by name of child service. A service is initially added to the map when it is
	// created via Add, and it is later removed or added to the map as needed,
	// through the SetStatus method.
	status map[string]struct{}

	// updaters is the list of hooks that are run (to update the Balancer parent(s)), whenever the Balancer status changes.
	// No mutex is needed, as it is modified only during the configuration build.
	updaters []func(bool)

	index            int
	currentWeight    int
	wantsHealthCheck bool
}

// NewWRRLoadBalancer creates a new WRRLoadBalancer.
func NewWRRLoadBalancer(wantsHealthCheck bool) *WRRLoadBalancer {
	return &WRRLoadBalancer{
		status:           make(map[string]struct{}),
		index:            -1,
		wantsHealthCheck: wantsHealthCheck,
	}
}

// ServeTCP forwards the connection to the right service.
func (b *WRRLoadBalancer) ServeTCP(conn WriteCloser) {
	next, err := b.nextServer()
	if err != nil {
		if !errors.Is(err, errNoServersInPool) {
			log.Error().Err(err).Msg("Error during load balancing")
		}
		_ = conn.Close()
		return
	}

	next.ServeTCP(conn)
}

// Add appends a server to the existing list with a name and weight.
func (b *WRRLoadBalancer) Add(name string, handler Handler, weight *int) {
	w := 1
	if weight != nil {
		w = *weight
	}

	b.serversMu.Lock()
	b.servers = append(b.servers, server{Handler: handler, name: name, weight: w})
	b.status[name] = struct{}{}
	b.serversMu.Unlock()
}

// SetStatus sets status (UP or DOWN) of a target server.
func (b *WRRLoadBalancer) SetStatus(ctx context.Context, childName string, up bool) {
	b.serversMu.Lock()
	defer b.serversMu.Unlock()

	upBefore := len(b.status) > 0

	status := "DOWN"
	if up {
		status = "UP"
	}

	log.Ctx(ctx).Debug().Msgf("Setting status of %s to %v", childName, status)

	if up {
		b.status[childName] = struct{}{}
	} else {
		delete(b.status, childName)
	}

	upAfter := len(b.status) > 0
	status = "DOWN"
	if upAfter {
		status = "UP"
	}

	// No Status Change
	if upBefore == upAfter {
		// We're still with the same status, no need to propagate
		log.Ctx(ctx).Debug().Msgf("Still %s, no need to propagate", status)
		return
	}

	// Status Change
	log.Ctx(ctx).Debug().Msgf("Propagating new %s status", status)
	for _, fn := range b.updaters {
		fn(upAfter)
	}
}

func (b *WRRLoadBalancer) RegisterStatusUpdater(fn func(up bool)) error {
	if !b.wantsHealthCheck {
		return errors.New("healthCheck not enabled in config for this weighted service")
	}

	b.updaters = append(b.updaters, fn)
	return nil
}

func (b *WRRLoadBalancer) nextServer() (Handler, error) {
	b.serversMu.Lock()
	defer b.serversMu.Unlock()

	if len(b.servers) == 0 || len(b.status) == 0 {
		return nil, errNoServersInPool
	}

	// The algo below may look messy, but is actually very simple
	// it calculates the GCD  and subtracts it on every iteration, what interleaves servers
	// and allows us not to build an iterator every time we readjust weights.

	// Maximum weight across all enabled servers.
	maximum := b.maxWeight()
	if maximum == 0 {
		return nil, errors.New("all servers have 0 weight")
	}

	// GCD across all enabled servers
	gcd := b.weightGcd()

	for {
		b.index = (b.index + 1) % len(b.servers)
		if b.index == 0 {
			b.currentWeight -= gcd
			if b.currentWeight <= 0 {
				b.currentWeight = maximum
			}
		}
		srv := b.servers[b.index]

		if _, ok := b.status[srv.name]; ok && srv.weight >= b.currentWeight {
			return srv, nil
		}
	}
}

func (b *WRRLoadBalancer) maxWeight() int {
	maximum := -1
	for _, s := range b.servers {
		if s.weight > maximum {
			maximum = s.weight
		}
	}
	return maximum
}

func (b *WRRLoadBalancer) weightGcd() int {
	divisor := -1
	for _, s := range b.servers {
		if divisor == -1 {
			divisor = s.weight
		} else {
			divisor = gcd(divisor, s.weight)
		}
	}
	return divisor
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
