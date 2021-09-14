package tcp

import (
	"fmt"
	"net"
	"strings"

	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/traefik/traefik/v2/pkg/types"
	"github.com/vulcand/predicate"
)

var funcs = map[string]func(*route, ...string) error{
	"HostSNI":  hostSNI,
	"ClientIP": clientIP,
}

// metaTCP contains TCP connection metadata
type metaTCP struct {
	serverName string
	remoteIP   string
}

// NewMetaTCP builds a metaTCP struct from the given parameters.
func NewMetaTCP(serverName string, conn WriteCloser) (metaTCP, error) {
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return metaTCP{}, fmt.Errorf("error while parsing remote address %q: %v", conn.RemoteAddr().String(), err)
	}

	return metaTCP{
		serverName: types.CanonicalDomain(serverName),
		remoteIP:   ip,
	}, nil
}

// TCPRouterMux handle TCP routing with rules.
type TCPRouterMux struct {
	subRouter
	parser predicate.Parser
}

func NewTCPRouterMux() (*TCPRouterMux, error) {
	parser, err := rules.NewTCPParser()
	if err != nil {
		return nil, fmt.Errorf("error while creating rules parser: %w", err)
	}

	return &TCPRouterMux{parser: parser}, nil
}

func (r TCPRouterMux) Match(meta metaTCP) Handler {
	// For each route, check if match, and return the handler for that route.
	for _, route := range r.routes {
		if route.match(meta) {
			return route.handler
		}
	}

	return nil
}

func (r *TCPRouterMux) HasRoutes() bool {
	return len(r.routes) > 0
}

// AddRoute add a new route to the router.
func (r *TCPRouterMux) AddRoute(rule string, handler Handler) error {
	parse, err := r.parser.Parse(rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return fmt.Errorf("error while parsing rule %s", rule)
	}

	route := &route{handler: handler}
	r.routes = append(r.routes, route)

	err = addRuleOnRoute(route, buildTree())
	if err != nil {
		route.buildOnly()
		return err
	}

	return nil
}

type subRouter struct {
	routes []*route
}

func (r *subRouter) newRoute() *route {
	route := &route{}
	r.routes = append(r.routes, route)
	return route
}

func (s subRouter) match(meta metaTCP) bool {
	// For each route, check if match, and return the handler for that route.
	for _, route := range s.routes {
		if route.match(meta) {
			return true
		}
	}

	return false
}

// matcher is a matcher used to match connection properties.
type matcher func(meta metaTCP) bool

// route holds matchers to match TCP routes.
type route struct {
	// List of matchers that will be used to match the route.
	matchers []matcher

	router *subRouter

	// Handler responsible for handling the route.
	handler Handler

	noMatch bool
}

func (r *route) buildOnly() {
	r.noMatch = true
}

func (r *route) subRouter() *subRouter {
	router := &subRouter{}
	r.router = router
	return router
}

// Match checks the connection against all the matchers in the route, and returns if there is a full match.
func (r *route) match(meta metaTCP) bool {
	if r.noMatch {
		return false
	}

	if len(r.matchers) == 0 && r.router != nil {
		return r.router.match(meta)
	}

	// For each matcher, check if match, and return true if all are matched.
	for _, matcher := range r.matchers {
		if !matcher(meta) {

			if r.router != nil {
				return r.router.match(meta)
			}

			return false
		}
	}

	// All matchers matched
	return true
}

// addMatcher adds a matcher to the route.
func (r *route) addMatcher(m matcher) {
	r.matchers = append(r.matchers, m)
}

func addRuleOnRouter(router *subRouter, rule *rules.Tree) error {
	switch rule.Matcher {
	case "and":
		route := router.newRoute()
		err := addRuleOnRoute(route, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRoute(route, rule.RuleRight)
	case "or":
		err := addRuleOnRouter(router, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRouter(router, rule.RuleRight)
	default:
		err := rules.CheckRule(rule)
		if err != nil {
			return err
		}

		if rule.Not {
			return not(funcs[rule.Matcher])(router.newRoute(), rule.Value...)
		}
		return funcs[rule.Matcher](router.newRoute(), rule.Value...)
	}
}

func addRuleOnRoute(route *route, rule *rules.Tree) error {
	switch rule.Matcher {
	case "and":
		err := addRuleOnRoute(route, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRoute(route, rule.RuleRight)
	case "or":
		subRouter := route.subRouter()

		err := addRuleOnRouter(subRouter, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRouter(subRouter, rule.RuleRight)
	default:
		err := rules.CheckRule(rule)
		if err != nil {
			return err
		}

		if rule.Not {
			return not(funcs[rule.Matcher])(route, rule.Value...)
		}
		return funcs[rule.Matcher](route, rule.Value...)
	}
}

func not(m func(*route, ...string) error) func(*route, ...string) error {
	return func(r *route, v ...string) error {
		router := TCPRouterMux{}
		err := m(router.newRoute(), v...)
		if err != nil {
			return err
		}

		r.addMatcher(func(meta metaTCP) bool {
			return !r.match(meta)
		})

		return nil
	}
}

func clientIP(route *route, clientIPs ...string) error {
	checker, err := ip.NewChecker(clientIPs)
	if err != nil {
		return fmt.Errorf("could not initialize IP Checker for \"ClientIP\" matcher: %w", err)
	}

	route.addMatcher(func(meta metaTCP) bool {
		if meta.remoteIP == "" {
			return false
		}

		ok, err := checker.Contains(meta.remoteIP)
		if err != nil {
			log.WithoutContext().Warnf("\"ClientIP\" matcher: could not match remote address : %w", err)
			return false
		}
		if ok {
			return true
		}

		return false
	})

	return nil
}

// hostSNI checks if the SNI Host of the connection match the matcher host.
func hostSNI(route *route, hosts ...string) error {
	route.addMatcher(func(meta metaTCP) bool {
		if len(hosts) == 0 {
			return false
		}
		// TODO verify if this is correct
		if hosts[0] == "*" {
			return true
		}

		if meta.serverName == "" {
			return false
		}

		for _, host := range hosts {
			if matchHost(meta.serverName, host) {
				return true
			}
		}

		return false
	})

	return nil
}

func matchHost(host, givenHost string) bool {
	if host == givenHost {
		return true
	}

	for len(givenHost) > 0 && givenHost[len(givenHost)-1] == '.' {
		givenHost = givenHost[:len(givenHost)-1]
	}

	labels := strings.Split(host, ".")
	for i := range labels {
		labels[i] = "*"
		candidate := strings.Join(labels, ".")
		if givenHost == candidate {
			return true
		}
	}

	return false
}
