// package roundrobin implements dynamic weighted round robin load balancer http handler
package roundrobin

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/vulcand/oxy/utils"
)

// Weight is an optional functional argument that sets weight of the server
func Weight(w int) ServerOption {
	return func(s *server) error {
		if w < 0 {
			return fmt.Errorf("Weight should be >= 0")
		}
		s.weight = w
		return nil
	}
}

// ErrorHandler is a functional argument that sets error handler of the server
func ErrorHandler(h utils.ErrorHandler) LBOption {
	return func(s *RoundRobin) error {
		s.errHandler = h
		return nil
	}
}

func EnableStickySession(ss *StickySession) LBOption {
	return func(s *RoundRobin) error {
		s.ss = ss
		return nil
	}
}

type RoundRobin struct {
	mutex      *sync.Mutex
	next       http.Handler
	errHandler utils.ErrorHandler
	// Current index (starts from -1)
	index         int
	servers       []*server
	currentWeight int
	ss            *StickySession
}

func New(next http.Handler, opts ...LBOption) (*RoundRobin, error) {
	rr := &RoundRobin{
		next:    next,
		index:   -1,
		mutex:   &sync.Mutex{},
		servers: []*server{},
		ss:      nil,
	}
	for _, o := range opts {
		if err := o(rr); err != nil {
			return nil, err
		}
	}
	if rr.errHandler == nil {
		rr.errHandler = utils.DefaultHandler
	}
	return rr, nil
}

func (r *RoundRobin) Next() http.Handler {
	return r.next
}

func (r *RoundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// make shallow copy of request before chaning anything to avoid side effects
	newReq := *req
	stuck := false
	if r.ss != nil {
		cookie_url, present, err := r.ss.GetBackend(&newReq, r.Servers())

		if err != nil {
			r.errHandler.ServeHTTP(w, req, err)
			return
		}

		if present {
			newReq.URL = cookie_url
			stuck = true
		}
	}

	if !stuck {
		url, err := r.NextServer()
		if err != nil {
			r.errHandler.ServeHTTP(w, req, err)
			return
		}

		if r.ss != nil {
			r.ss.StickBackend(url, &w)
		}
		newReq.URL = url
	}
	r.next.ServeHTTP(w, &newReq)
}

func (r *RoundRobin) NextServer() (*url.URL, error) {
	srv, err := r.nextServer()
	if err != nil {
		return nil, err
	}
	return utils.CopyURL(srv.url), nil
}

func (r *RoundRobin) nextServer() (*server, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.servers) == 0 {
		return nil, fmt.Errorf("no servers in the pool")
	}

	// The algo below may look messy, but is actually very simple
	// it calculates the GCD  and subtracts it on every iteration, what interleaves servers
	// and allows us not to build an iterator every time we readjust weights

	// GCD across all enabled servers
	gcd := r.weightGcd()
	// Maximum weight across all enabled servers
	max := r.maxWeight()

	for {
		r.index = (r.index + 1) % len(r.servers)
		if r.index == 0 {
			r.currentWeight = r.currentWeight - gcd
			if r.currentWeight <= 0 {
				r.currentWeight = max
				if r.currentWeight == 0 {
					return nil, fmt.Errorf("all servers have 0 weight")
				}
			}
		}
		srv := r.servers[r.index]
		if srv.weight >= r.currentWeight {
			return srv, nil
		}
	}
	// We did full circle and found no available servers
	return nil, fmt.Errorf("no available servers")
}

func (r *RoundRobin) RemoveServer(u *url.URL) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	e, index := r.findServerByURL(u)
	if e == nil {
		return fmt.Errorf("server not found")
	}
	r.servers = append(r.servers[:index], r.servers[index+1:]...)
	r.resetState()
	return nil
}

func (rr *RoundRobin) Servers() []*url.URL {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	out := make([]*url.URL, len(rr.servers))
	for i, srv := range rr.servers {
		out[i] = srv.url
	}
	return out
}

func (rr *RoundRobin) ServerWeight(u *url.URL) (int, bool) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if s, _ := rr.findServerByURL(u); s != nil {
		return s.weight, true
	}
	return -1, false
}

// In case if server is already present in the load balancer, returns error
func (rr *RoundRobin) UpsertServer(u *url.URL, options ...ServerOption) error {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if u == nil {
		return fmt.Errorf("server URL can't be nil")
	}

	if s, _ := rr.findServerByURL(u); s != nil {
		for _, o := range options {
			if err := o(s); err != nil {
				return err
			}
		}
		rr.resetState()
		return nil
	}

	srv := &server{url: utils.CopyURL(u)}
	for _, o := range options {
		if err := o(srv); err != nil {
			return err
		}
	}

	if srv.weight == 0 {
		srv.weight = defaultWeight
	}

	rr.servers = append(rr.servers, srv)
	rr.resetState()
	return nil
}

func (r *RoundRobin) resetIterator() {
	r.index = -1
	r.currentWeight = 0
}

func (r *RoundRobin) resetState() {
	r.resetIterator()
}

func (r *RoundRobin) findServerByURL(u *url.URL) (*server, int) {
	if len(r.servers) == 0 {
		return nil, -1
	}
	for i, s := range r.servers {
		if sameURL(u, s.url) {
			return s, i
		}
	}
	return nil, -1
}

func (rr *RoundRobin) maxWeight() int {
	max := -1
	for _, s := range rr.servers {
		if s.weight > max {
			max = s.weight
		}
	}
	return max
}

func (rr *RoundRobin) weightGcd() int {
	divisor := -1
	for _, s := range rr.servers {
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

// ServerOption provides various options for server, e.g. weight
type ServerOption func(*server) error

// LBOption provides options for load balancer
type LBOption func(*RoundRobin) error

// Set additional parameters for the server can be supplied when adding server
type server struct {
	url *url.URL
	// Relative weight for the enpoint to other enpoints in the load balancer
	weight int
}

const defaultWeight = 1

func sameURL(a, b *url.URL) bool {
	return a.Path == b.Path && a.Host == b.Host && a.Scheme == b.Scheme
}

type balancerHandler interface {
	Servers() []*url.URL
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	ServerWeight(u *url.URL) (int, bool)
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...ServerOption) error
	NextServer() (*url.URL, error)
	Next() http.Handler
}
