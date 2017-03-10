/*
package route provides http package-compatible routing library. It can route http requests by by hostname, method, path and headers.

Route defines simple language for matching requests based on Go syntax. Route provides series of matchers that follow the syntax:


   Matcher("value")          // matches value using trie
   Matcher("<string>.value") // uses trie-based matching for a.value and b.value
   MatcherRegexp(".*value")  // uses regexp-based matching

Host matcher:

  Host("<subdomain>.localhost") // trie-based matcher for a.localhost, b.localhost, etc.
  HostRegexp(".*localhost")     // regexp based matcher

Path matcher:

  Path("/hello/<value>")   // trie-based matcher for raw request path
  PathRegexp("/hello/.*")  // regexp-based matcher for raw request path

Method matcher:

  Method("GET")            // trie-based matcher for request method
  MethodRegexp("POST|PUT") // regexp based matcher for request method

Header matcher:

  Header("Content-Type", "application/<subtype>") // trie-based matcher for headers
  HeaderRegexp("Content-Type", "application/.*")  // regexp based matcher for headers

Matchers can be combined using && operator:

  Host("localhost") && Method("POST") && Path("/v1")

Route library will join the trie-based matchers into one trie matcher when possible, for example:

  Host("localhost") && Method("POST") && Path("/v1")
  Host("localhost") && Method("GET") && Path("/v2")

Will be combined into one trie for performance. If you add a third route:

  Host("localhost") && Method("GET") && PathRegexp("/v2/.*")

It wont be joined ito the trie, and would be matched separatedly instead.
*/
package route

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
)

// Router implements http request routing and operations. It is a generic router not conforming to http.Handler interface, to get a handler
// conforming to http.Handler interface, use Mux router instead.
type Router interface {
	// GetRoute returns a route by a given expression, returns nil if expresison is not found
	GetRoute(string) interface{}

	// AddRoute adds a route to match by expression, returns error if the expression already defined, or route expression is incorrect
	AddRoute(string, interface{}) error

	// RemoveRoute removes a route for a given expression
	RemoveRoute(string) error

	// UpsertRoute updates an existing route or adds a new route by given expression
	UpsertRoute(string, interface{}) error

	// Route takes a request and matches it against requests, returns matched route in case if found, nil if there's no matching route or error in case of internal error.
	Route(*http.Request) (interface{}, error)
}

type router struct {
	mutex    *sync.RWMutex
	matchers []matcher
	routes   map[string]*match
}

// New creates a new Router instance
func New() Router {
	return &router{
		mutex:  &sync.RWMutex{},
		routes: make(map[string]*match),
	}
}

func (e *router) GetRoute(expr string) interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	res, ok := e.routes[expr]
	if ok {
		return res.val
	}
	return nil
}

func (e *router) AddRoute(expr string, val interface{}) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, ok := e.routes[expr]; ok {
		return fmt.Errorf("Expression '%s' already exists", expr)
	}
	result := &match{val: val}
	if _, err := parse(expr, result); err != nil {
		return err
	}
	e.routes[expr] = result
	if err := e.compile(); err != nil {
		delete(e.routes, expr)
		return err
	}
	return nil
}

func (e *router) UpsertRoute(expr string, val interface{}) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	result := &match{val: val}
	if _, err := parse(expr, result); err != nil {
		return err
	}
	prev, existed := e.routes[expr]

	e.routes[expr] = result
	if err := e.compile(); err != nil {
		if existed {
			e.routes[expr] = prev
		} else {
			delete(e.routes, expr)
		}
		return err
	}
	return nil
}

func (e *router) compile() error {
	var exprs = []string{}
	for expr, _ := range e.routes {
		exprs = append(exprs, expr)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(exprs)))

	matchers := []matcher{}
	i := 0
	for _, expr := range exprs {
		result := e.routes[expr]
		matcher, err := parse(expr, result)
		if err != nil {
			return err
		}

		// Merge the previous and new matcher if that's possible
		if i > 0 && matchers[i-1].canMerge(matcher) {
			m, err := matchers[i-1].merge(matcher)
			if err != nil {
				return err
			}
			matchers[i-1] = m
		} else {
			matchers = append(matchers, matcher)
			i += 1
		}
	}

	e.matchers = matchers
	return nil
}

func (e *router) RemoveRoute(expr string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	delete(e.routes, expr)
	return e.compile()
}

func (e *router) Route(req *http.Request) (interface{}, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if len(e.matchers) == 0 {
		return nil, nil
	}

	for _, m := range e.matchers {
		if l := m.match(req); l != nil {
			return l.val, nil
		}
	}
	return nil, nil
}
