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
	weight int
}

// WRRLoadBalancer is a naive RoundRobin load balancer for TCP services.
type WRRLoadBalancer struct {
	serversMu sync.Mutex
	servers   []server

	// status is a record of which child services of the Balancer are healthy.
	status sync.Map

	// updaters is the list of hooks that are run (to update the Balancer
	// parent(s)), whenever the Balancer status changes.
	// Modified during configuration build only, so no mutex needed.
	updaters []func(bool)

	currentWeight    int
	index            int
	wantsHealthCheck bool
}

// NewWRRLoadBalancer creates a new WRRLoadBalancer.
func NewWRRLoadBalancer(wantsHealthCheck bool) *WRRLoadBalancer {
	return &WRRLoadBalancer{
		wantsHealthCheck: wantsHealthCheck,
		index:            -1,
	}
}

// ServeTCP forwards the connection to the right service.
func (b *WRRLoadBalancer) ServeTCP(conn WriteCloser) {
	b.serversMu.Lock()
	next, err := b.next()
	b.serversMu.Unlock()

	if err != nil {
		if !errors.Is(err, errNoServersInPool) {
			log.Error().Err(err).Msg("Error during load balancing")
		}
		_ = conn.Close()
		return
	}

	next.ServeTCP(conn)
}

// AddServer appends a server to the existing list.
func (b *WRRLoadBalancer) AddServer(serverHandler Handler) {
	w := 1
	b.AddWeightServer(serverHandler, &w)
}

// AddWeightServer appends a server to the existing list with a weight.
func (b *WRRLoadBalancer) AddWeightServer(serverHandler Handler, weight *int) {
	b.serversMu.Lock()
	defer b.serversMu.Unlock()

	w := 1
	if weight != nil {
		w = *weight
	}
	b.servers = append(b.servers, server{Handler: serverHandler, weight: w})
}

// SetStatus sets status (UP or DOWN) of a target server.
func (b *WRRLoadBalancer) SetStatus(ctx context.Context, childName string, up bool) {
	statusString := "DOWN"
	if up {
		statusString = "UP"
	}

	log.Ctx(ctx).Debug().Msgf("Setting status of %s to %s", childName, statusString)

	if _, loaded := b.status.LoadOrStore(childName, up); !loaded {
		log.Ctx(ctx).Debug().Msgf("Propagating new %s status", statusString)

		for _, fn := range b.updaters {
			fn(up)
		}

		return
	}

	log.Ctx(ctx).Debug().Msgf("Still %s, no need to propagate", statusString)
}

func (b *WRRLoadBalancer) RegisterStatusUpdater(fn func(up bool)) error {
	if !b.wantsHealthCheck {
		return errors.New("healthCheck not enabled in config for this weighted service")
	}

	b.updaters = append(b.updaters, fn)
	return nil
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

func (b *WRRLoadBalancer) next() (Handler, error) {
	if len(b.servers) == 0 {
		return nil, errNoServersInPool
	}

	// The algo below may look messy, but is actually very simple
	// it calculates the GCD  and subtracts it on every iteration, what interleaves servers
	// and allows us not to build an iterator every time we readjust weights

	// Maximum weight across all enabled servers
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
		if srv.weight >= b.currentWeight {
			return srv, nil
		}
	}
}
