package wrr

import (
	"container/heap"
	"fmt"
	"net/http"
	"sync"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
)

type namedHandler struct {
	http.Handler
	name     string
	weight   float64
	deadline float64
}

type stickyCookie struct {
	name     string
	secure   bool
	httpOnly bool
}

// New creates a new load balancer.
func New(sticky *dynamic.Sticky) *Balancer {
	handlers := make(namedHandlers, 0)
	balancer := &Balancer{
		handlers: &handlers,
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

type namedHandlers []*namedHandler

// Len implements heap.Interface/sort.Interface
func (n namedHandlers) Len() int { return len(n) }

// Less implements heap.Interface/sort.Interface
func (n namedHandlers) Less(i, j int) bool {
	return n[i].deadline < n[j].deadline
}

// Swap implements heap.Interface/sort.Interface
func (n namedHandlers) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// Push implements heap.Interface for pushing an item into the heap
func (n *namedHandlers) Push(x interface{}) {
	h, ok := x.(*namedHandler)
	if !ok {
		return
	}
	*n = append(*n, h)
}

// Pop implements heap.Interface for poping an item from the heap
func (n *namedHandlers) Pop() interface{} {
	old := *n
	h := old[len(old)-1]
	*n = old[0 : len(old)-1]
	return h
}

// Balancer is a WeightedRoundRobin load balancer based on Earliest Deadline First (EDF).
// (https://en.wikipedia.org/wiki/Earliest_deadline_first_scheduling)
// Each pick from the schedule has the earliest deadline entry selected. Entries have deadlines set
// at currentDeadline + 1 / weight, providing weighted round robin behavior with floating point
// weights and an O(log n) pick time.
type Balancer struct {
	stickyCookie *stickyCookie

	mutex       sync.RWMutex
	handlers    *namedHandlers
	curDeadline float64
}

func (b *Balancer) nextServer() (*namedHandler, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if b.handlers == nil || len(*b.handlers) == 0 {
		return nil, fmt.Errorf("no servers in the pool")
	}
	// Pick handler with closest deadline.
	handler := heap.Pop(b.handlers).(*namedHandler)
	// curDeadline should be handler's deadline so that new added entry would have a fair
	// competition environment with the old ones.
	b.curDeadline = handler.deadline
	handler.deadline += 1 / handler.weight
	heap.Push(b.handlers, handler)
	log.WithoutContext().Debugf("Service selected by WRR: %s", handler.name)
	return handler, nil
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if b.stickyCookie != nil {
		cookie, err := req.Cookie(b.stickyCookie.name)

		if err != nil && err != http.ErrNoCookie {
			log.WithoutContext().Warnf("Error while reading cookie: %v", err)
		}

		if err == nil && cookie != nil {
			for _, handler := range *b.handlers {
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
// It is not thread safe with ServeHTTP.
func (b *Balancer) AddService(name string, handler http.Handler, weight *int) {
	w := 1
	if weight != nil {
		w = *weight
	}
	if w <= 0 { // non-positive weight is meaningless
		return
	}
	h := &namedHandler{Handler: handler, name: name, weight: float64(w)}
	// use RWLock to protect b.curDeadline
	b.mutex.RLock()
	h.deadline = b.curDeadline + 1/h.weight
	b.mutex.RUnlock()
	heap.Push(b.handlers, h)
}
