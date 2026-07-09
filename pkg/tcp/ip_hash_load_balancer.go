package tcp

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	statusUp   = "UP"
	statusDown = "DOWN"
)

type IPHashLoadBalancer struct {
	serversMu sync.RWMutex
	servers   []server
	// status is a record of which child services of the Balancer are healthy, keyed
	// by name of child service. A service is initially added to the map when it is
	// created via Add, and it is later removed or added to the map as needed,
	// through the SetStatus method.
	status map[string]struct{}

	// updaters is the list of hooks that are run (to update the Balancer parent(s)), whenever the Balancer status changes.
	// No mutex is needed, as it is modified only during the configuration build.
	updaters []func(bool)

	wantsHealthCheck bool
}

func NewIPHashLoadBalancer(wantsHealthCheck bool) *IPHashLoadBalancer {
	return &IPHashLoadBalancer{
		status:           make(map[string]struct{}),
		wantsHealthCheck: wantsHealthCheck,
	}
}

// ServeTCP forwards the connection to the right service.
func (b *IPHashLoadBalancer) ServeTCP(conn WriteCloser) {
	clientIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Error().Err(err).Msg("IP hash load balancer: failed to parse client address")
		_ = conn.Close()
		return
	}

	next, err := b.nextServer(clientIP)
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
func (b *IPHashLoadBalancer) Add(name string, handler Handler, weight *int) {
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
func (b *IPHashLoadBalancer) SetStatus(ctx context.Context, childName string, up bool) {
	b.serversMu.Lock()
	defer b.serversMu.Unlock()

	upBefore := len(b.status) > 0

	status := statusDown
	if up {
		status = statusUp
	}

	log.Ctx(ctx).Debug().Msgf("Setting status of %s to %v", childName, status)

	if up {
		b.status[childName] = struct{}{}
	} else {
		delete(b.status, childName)
	}

	upAfter := len(b.status) > 0
	status = statusDown
	if upAfter {
		status = statusUp
	}

	// No Status Change
	if upBefore == upAfter {
		log.Ctx(ctx).Debug().Msgf("Still %s, no need to propagate", status)
		return
	}

	// Status Change
	log.Ctx(ctx).Debug().Msgf("Propagating new %s status", status)
	for _, fn := range b.updaters {
		fn(upAfter)
	}
}

func (b *IPHashLoadBalancer) RegisterStatusUpdater(fn func(up bool)) error {
	if !b.wantsHealthCheck {
		return errors.New("healthCheck not enabled in config for this IP hash service")
	}

	b.updaters = append(b.updaters, fn)
	return nil
}

func (b *IPHashLoadBalancer) nextServer(clientIP string) (Handler, error) {
	b.serversMu.RLock()
	defer b.serversMu.RUnlock()

	var healthyServers []server
	for _, server := range b.servers {
		if _, ok := b.status[server.name]; ok {
			healthyServers = append(healthyServers, server)
		}
	}

	if len(healthyServers) == 0 {
		return nil, errNoServersInPool
	}

	maxScore := math.Inf(-1)
	var serverSelected Handler
	for _, s := range healthyServers {
		score := ipHashNodeScore(s.name, clientIP, s.weight)
		if score > maxScore {
			maxScore = score
			serverSelected = s.Handler
		}
	}

	return serverSelected, nil
}

func ipHashNodeScore(serverName, clientIP string, weight int) float64 {
	h := fnv.New64a()
	h.Write([]byte(clientIP + serverName))
	sum := h.Sum64()
	score := float64(sum) / math.Pow(2, 64)
	logScore := 1.0 / -math.Log(score)

	return logScore * float64(weight)
}
