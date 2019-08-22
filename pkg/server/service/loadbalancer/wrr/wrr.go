package wrr

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
)

type namedHandler struct {
	http.Handler
	name   string
	weight int
}

type stickyCookie struct {
	name     string
	secure   bool
	httpOnly bool
}

// New creates a new load balancer.
func New(sticky *dynamic.Sticky) *Balancer {
	balancer := &Balancer{
		mutex: &sync.Mutex{},
		index: -1,
	}
	if sticky != nil && sticky.Cookie != nil {
		balancer.stickyCookie = &stickyCookie{
			name:     sticky.Cookie.Name,
			secure:   sticky.Cookie.Secure,
			httpOnly: sticky.Cookie.HTTPOnly,
		}
	}
	return balancer
}

// Balancer is a WeightedRoundRobin load balancer.
type Balancer struct {
	handlers []*namedHandler
	mutex    *sync.Mutex
	// Current index (starts from -1)
	index         int
	currentWeight int
	stickyCookie  *stickyCookie
}

func (b *Balancer) maxWeight() int {
	max := -1
	for _, s := range b.handlers {
		if s.weight > max {
			max = s.weight
		}
	}
	return max
}

func (b *Balancer) weightGcd() int {
	divisor := -1
	for _, s := range b.handlers {
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

func (b *Balancer) nextServer() (*namedHandler, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.handlers) == 0 {
		return nil, fmt.Errorf("no servers in the pool")
	}

	// The algo below may look messy, but is actually very simple
	// it calculates the GCD  and subtracts it on every iteration, what interleaves servers
	// and allows us not to build an iterator every time we readjust weights

	// GCD across all enabled servers
	gcd := b.weightGcd()
	// Maximum weight across all enabled servers
	max := b.maxWeight()

	for {
		b.index = (b.index + 1) % len(b.handlers)
		if b.index == 0 {
			b.currentWeight -= gcd
			if b.currentWeight <= 0 {
				b.currentWeight = max
				if b.currentWeight == 0 {
					return nil, fmt.Errorf("all servers have 0 weight")
				}
			}
		}
		srv := b.handlers[b.index]
		if srv.weight >= b.currentWeight {
			log.WithoutContext().Debugf("Service Select: %s", srv.name)
			return srv, nil
		}
	}
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if b.stickyCookie != nil {
		cookie, err := req.Cookie(b.stickyCookie.name)

		if err != nil && err != http.ErrNoCookie {
			log.WithoutContext().Warnf("Error while reading cookie: %v", err)
		}

		if err == nil && cookie != nil {
			for _, handler := range b.handlers {
				if handler.name == cookie.Value {
					handler.ServeHTTP(w, req)
					return
				}
			}
		}
	}

	server, err := b.nextServer()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError)+err.Error(), http.StatusInternalServerError)
		return
	}

	if b.stickyCookie != nil {
		cookie := &http.Cookie{Name: b.stickyCookie.name, Value: server.name, Path: "/", HttpOnly: b.stickyCookie.httpOnly, Secure: b.stickyCookie.secure}
		http.SetCookie(w, cookie)
	}

	server.ServeHTTP(w, req)
}

// AddService adds a handler.
func (b *Balancer) AddService(name string, handler http.Handler, weight *int) {
	w := 1
	if weight != nil {
		w = *weight
	}
	b.handlers = append(b.handlers, &namedHandler{Handler: handler, name: name, weight: w})
}
