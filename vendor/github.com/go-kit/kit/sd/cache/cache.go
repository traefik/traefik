package cache

import (
	"io"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
)

// Cache collects the most recent set of endpoints from a service discovery
// system via a subscriber, and makes them available to consumers. Cache is
// meant to be embedded inside of a concrete subscriber, and can serve Service
// invocations directly.
type Cache struct {
	mtx     sync.RWMutex
	factory sd.Factory
	cache   map[string]endpointCloser
	slice   atomic.Value // []endpoint.Endpoint
	logger  log.Logger
}

type endpointCloser struct {
	endpoint.Endpoint
	io.Closer
}

// New returns a new, empty endpoint cache.
func New(factory sd.Factory, logger log.Logger) *Cache {
	return &Cache{
		factory: factory,
		cache:   map[string]endpointCloser{},
		logger:  logger,
	}
}

// Update should be invoked by clients with a complete set of current instance
// strings whenever that set changes. The cache manufactures new endpoints via
// the factory, closes old endpoints when they disappear, and persists existing
// endpoints if they survive through an update.
func (c *Cache) Update(instances []string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Deterministic order (for later).
	sort.Strings(instances)

	// Produce the current set of services.
	cache := make(map[string]endpointCloser, len(instances))
	for _, instance := range instances {
		// If it already exists, just copy it over.
		if sc, ok := c.cache[instance]; ok {
			cache[instance] = sc
			delete(c.cache, instance)
			continue
		}

		// If it doesn't exist, create it.
		service, closer, err := c.factory(instance)
		if err != nil {
			c.logger.Log("instance", instance, "err", err)
			continue
		}
		cache[instance] = endpointCloser{service, closer}
	}

	// Close any leftover endpoints.
	for _, sc := range c.cache {
		if sc.Closer != nil {
			sc.Closer.Close()
		}
	}

	// Populate the slice of endpoints.
	slice := make([]endpoint.Endpoint, 0, len(cache))
	for _, instance := range instances {
		// A bad factory may mean an instance is not present.
		if _, ok := cache[instance]; !ok {
			continue
		}
		slice = append(slice, cache[instance].Endpoint)
	}

	// Swap and trigger GC for old copies.
	c.slice.Store(slice)
	c.cache = cache
}

// Endpoints yields the current set of (presumably identical) endpoints, ordered
// lexicographically by the corresponding instance string.
func (c *Cache) Endpoints() []endpoint.Endpoint {
	return c.slice.Load().([]endpoint.Endpoint)
}
