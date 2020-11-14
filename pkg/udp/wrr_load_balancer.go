package udp

import (
	"fmt"
	"sync"

	"github.com/traefik/traefik/v2/pkg/log"
)

type server struct {
	Handler
	weight int
}

// WRRLoadBalancer is a naive RoundRobin load balancer for UDP services.
type WRRLoadBalancer struct {
	servers       []server
	lock          sync.RWMutex
	currentWeight int
	index         int
}

// NewWRRLoadBalancer creates a new WRRLoadBalancer.
func NewWRRLoadBalancer() *WRRLoadBalancer {
	return &WRRLoadBalancer{
		index: -1,
	}
}

// ServeUDP forwards the connection to the right service.
func (b *WRRLoadBalancer) ServeUDP(conn *Conn) {
	if len(b.servers) == 0 {
		log.WithoutContext().Error("no available server")
		return
	}

	next, err := b.next()
	if err != nil {
		log.WithoutContext().Errorf("Error during load balancing: %v", err)
		conn.Close()
	}
	next.ServeUDP(conn)
}

// AddServer appends a handler to the existing list.
func (b *WRRLoadBalancer) AddServer(serverHandler Handler) {
	w := 1
	b.AddWeightedServer(serverHandler, &w)
}

// AddWeightedServer appends a handler to the existing list with a weight.
func (b *WRRLoadBalancer) AddWeightedServer(serverHandler Handler, weight *int) {
	w := 1
	if weight != nil {
		w = *weight
	}
	b.servers = append(b.servers, server{Handler: serverHandler, weight: w})
}

func (b *WRRLoadBalancer) maxWeight() int {
	max := -1
	for _, s := range b.servers {
		if s.weight > max {
			max = s.weight
		}
	}
	return max
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
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.servers) == 0 {
		return nil, fmt.Errorf("no servers in the pool")
	}

	// The algorithm below may look messy,
	// but is actually very simple it calculates the GCD  and subtracts it on every iteration,
	// what interleaves servers and allows us not to build an iterator every time we readjust weights.

	// GCD across all enabled servers
	gcd := b.weightGcd()
	// Maximum weight across all enabled servers
	max := b.maxWeight()

	for {
		b.index = (b.index + 1) % len(b.servers)
		if b.index == 0 {
			b.currentWeight -= gcd
			if b.currentWeight <= 0 {
				b.currentWeight = max
				if b.currentWeight == 0 {
					return nil, fmt.Errorf("all servers have 0 weight")
				}
			}
		}
		srv := b.servers[b.index]
		if srv.weight >= b.currentWeight {
			return srv, nil
		}
	}
}
