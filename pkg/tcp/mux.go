package tcp

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"sort"
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

// ParseHostSNI extracts the HostSNIs declared in a rule.
// This is a first naive implementation used in TCP routing.
func ParseHostSNI(rule string) ([]string, error) {
	var matchers []string
	for matcher := range funcs {
		matchers = append(matchers, matcher)
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		return nil, err
	}

	parse, err := parser.Parse(rule)
	if err != nil {
		return nil, err
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return nil, errors.New("cannot parse")
	}

	return buildTree().ParseMatchers([]string{"HostSNI"}), nil
}

// ConnData contains TCP connection metadata.
type ConnData struct {
	serverName string
	remoteIP   string
}

// NewConnData builds a connData struct from the given parameters.
func NewConnData(serverName string, conn WriteCloser) (ConnData, error) {
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return ConnData{}, fmt.Errorf("error while parsing remote address %q: %w", conn.RemoteAddr().String(), err)
	}

	// as per https://datatracker.ietf.org/doc/html/rfc6066:
	// > The hostname is represented as a byte string using ASCII encoding without a trailing dot.
	// so there is no need to trim a potential trailing dot
	serverName = types.CanonicalDomain(serverName)

	return ConnData{
		serverName: types.CanonicalDomain(serverName),
		remoteIP:   ip,
	}, nil
}

// Muxer defines a muxer that handles TCP routing with rules.
type Muxer struct {
	subRouter
	parser predicate.Parser
}

// NewMuxer returns a TCP muxer.
func NewMuxer() (*Muxer, error) {
	var matchers []string
	for matcher := range funcs {
		matchers = append(matchers, matcher)
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		return nil, fmt.Errorf("error while creating rules parser: %w", err)
	}

	return &Muxer{parser: parser}, nil
}

// Match returns the handler of the first route matching the connection metadata.
func (m Muxer) Match(meta ConnData) Handler {
	for _, route := range m.routes {
		if route.match(meta) {
			return route.handler
		}
	}

	return nil
}

// AddRoute adds a new route to the router.
func (m *Muxer) AddRoute(rule string, priority int, handler Handler) error {
	// Special case for catchAll rule
	// We set the lowest computable priority minus one to this handler,
	// in order to make it the last to be evaluated.
	if priority == 0 && rule == "HostSNI(`*`)" {
		priority = -1
	}

	// Default value, user has not set it, so we'll compute it.
	if priority == 0 {
		priority = len(rule)
	}

	parse, err := m.parser.Parse(rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return fmt.Errorf("error while parsing rule %s", rule)
	}

	newRoute := m.newRoute()
	newRoute.handler = handler
	newRoute.priority = priority

	err = addRuleOnRoute(newRoute, buildTree())
	if err != nil {
		newRoute.buildOnly()
		return err
	}

	sort.Sort(routes(m.routes))

	return nil
}

type routes []*route

func (r routes) Len() int      { return len(r) }
func (r routes) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r routes) Less(i, j int) bool {
	return r[i].priority > r[j].priority
}

func (m *Muxer) hasRoutes() bool {
	return len(m.routes) > 0
}

type subRouter struct {
	routes []*route
}

func (s *subRouter) newRoute() *route {
	route := &route{}
	s.routes = append(s.routes, route)
	return route
}

func (s subRouter) match(meta ConnData) bool {
	// For each route, check if match, and return the handler for that route.
	for _, route := range s.routes {
		if route.match(meta) {
			return true
		}
	}

	return false
}

// matcher is a matcher used to match connection properties.
type matcher func(meta ConnData) bool

// route holds matchers to match TCP routes.
type route struct {
	// List of matchers that will be used to match the route.
	matchers []matcher

	router *subRouter

	// Handler responsible for handling the route.
	handler Handler

	// TODO: doc + type
	priority int

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

// match checks the connection against all the matchers in the route, and returns if there is a full match.
func (r *route) match(meta ConnData) bool {
	if r.noMatch {
		return false
	}

	if len(r.matchers) == 0 && r.router != nil {
		return r.router.match(meta)
	}

	// For each matcher, check if it matches, and return true if all are matched.
	for _, matcher := range r.matchers {
		if !matcher(meta) {
			// TODO check why this is not covered by unit tests
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
		router := subRouter{}
		err := m(router.newRoute(), v...)
		if err != nil {
			return err
		}

		r.addMatcher(func(meta ConnData) bool {
			return !router.match(meta)
		})

		return nil
	}
}

func clientIP(route *route, clientIPs ...string) error {
	checker, err := ip.NewChecker(clientIPs)
	if err != nil {
		return fmt.Errorf("could not initialize IP Checker for \"ClientIP\" matcher: %w", err)
	}

	route.addMatcher(func(meta ConnData) bool {
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

var almostFQDN = regexp.MustCompile(`^[[:alnum:]\.-]+$`)

// hostSNI checks if the SNI Host of the connection match the matcher host.
func hostSNI(route *route, hosts ...string) error {
	if len(hosts) == 0 {
		return fmt.Errorf("empty value for \"HostSNI\" matcher is not allowed")
	}

	for i, host := range hosts {
		// Special case to allow global wildcard
		if host == "*" {
			continue
		}

		if !almostFQDN.MatchString(host) {
			return fmt.Errorf("invalid value for \"HostSNI\" matcher, %q is not a valid hostname", host)
		}

		hosts[i] = strings.ToLower(host)
	}

	route.addMatcher(func(meta ConnData) bool {
		// As HostSNI(`*`) rule has been prodided as catchAll for non-TLS TCP,
		// we allow matching with an empty serverName only for that case.
		if len(hosts) == 1 && hosts[0] == "*" {
			return true
		}

		if meta.serverName == "" {
			return false
		}

		for _, host := range hosts {
			if host == "*" {
				return true
			}

			if host == meta.serverName {
				return true
			}

			// trim trailing period in case of FQDN
			host = strings.TrimSuffix(host, ".")
			if host == meta.serverName {
				return true
			}
		}

		return false
	})

	return nil
}
