package tcp

import (
	"net"
	"sync"

	"github.com/containous/traefik/log"
)

// RRLoadBalancer is a naive RoundRobin load balancer for TCP services
type RRLoadBalancer struct {
	servers []Handler
	lock    sync.RWMutex
	current int
}

// NewRRLoadBalancer creates a new RRLoadBalancer
func NewRRLoadBalancer() *RRLoadBalancer {
	return &RRLoadBalancer{}
}

// ServeTCP forwards the connection to the right service
func (r *RRLoadBalancer) ServeTCP(conn net.Conn) {
	r.next().ServeTCP(conn)
}

// AddServer appends a server to the existing list
func (r *RRLoadBalancer) AddServer(server Handler) {
	r.servers = append(r.servers, server)
}

func (r *RRLoadBalancer) next() Handler {
	r.lock.Lock()
	defer r.lock.Unlock()

	// FIXME handle weight
	if r.current >= len(r.servers) {
		r.current = 0
		log.Debugf("Load balancer: going back to the first available server")
	}
	handler := r.servers[r.current]
	r.current++
	return handler
}
