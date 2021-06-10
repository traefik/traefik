package tcpmux

import (
	"github.com/traefik/traefik/v2/pkg/tcp"
)

type Route struct {
	// List of matchers that will be used to match the route.
	matchers []Matcher
	// Handler responsible for handling the route.
	handler tcp.Handler
}

// NewRoute returns a new empty Route.
func NewRoute() *Route {
	return &Route{}
}

// Match checks the connection against all the matchers in the route, and returns if there is a full match.
func (r *Route) Match(conn tcp.WriteCloser) bool {
	// For each matcher, check if match, and return true if all are matched.
	for _, matcher := range r.matchers {
		if !matcher.Match(conn) {
			return false
		}
	}
	// All matchers matched
	return true
}

// AddMatcher adds a matcher to the route.
func (r *Route) AddMatcher(m Matcher) *Route {
	r.matchers = append(r.matchers, m)
	return r
}
