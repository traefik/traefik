package tcp

import (
	"errors"
	"fmt"
	"net"
	"regexp"
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

// connData contains TCP connection metadata.
type connData struct {
	serverName string
	remoteIP   string
}

// NewConnData builds a connData struct from the given parameters.
func NewConnData(serverName string, conn WriteCloser) (connData, error) {
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return connData{}, fmt.Errorf("error while parsing remote address %q: %w", conn.RemoteAddr().String(), err)
	}

	// as per https://datatracker.ietf.org/doc/html/rfc6066:
	// > The hostname is represented as a byte string using ASCII encoding without a trailing dot.
	// so there is no need to trim a potential trailing dot
	serverName = types.CanonicalDomain(serverName)

	return connData{
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
func (r Muxer) Match(meta connData) Handler {
	// For each route, check if match, and return the handler for that route.
	for _, route := range r.routes {
		if route.match(meta) {
			return route.handler
		}
	}

	return nil
}

// AddRoute add a new route to the router.
func (r *Muxer) AddRoute(rule string, handler Handler) error {
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

func (s *subRouter) newRoute() *route {
	route := &route{}
	s.routes = append(s.routes, route)
	return route
}

func (s subRouter) match(meta connData) bool {
	// For each route, check if match, and return the handler for that route.
	for _, route := range s.routes {
		if route.match(meta) {
			return true
		}
	}

	return false
}

// matcher is a matcher used to match connection properties.
type matcher func(meta connData) bool

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
func (r *route) match(meta connData) bool {
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
		router := Muxer{}
		err := m(router.newRoute(), v...)
		if err != nil {
			return err
		}

		r.addMatcher(func(meta connData) bool {
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

	route.addMatcher(func(meta connData) bool {
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
		return fmt.Errorf("empty value for \"Host\" matcher is not allowed")
	}

	for i, host := range hosts {
		// Special case to allow global wildcard
		if host == "*" {
			hosts[i] = strings.ToLower(host)
			continue
		}

		if !almostFQDN.MatchString(host) {
			return fmt.Errorf("invalid value for \"HostSNI\" matcher, %q is not a valid hostname", host)
		}

		hosts[i] = strings.ToLower(host)
	}

	route.addMatcher(func(meta connData) bool {
		// TODO verify if this is correct
		if hosts[0] == "*" {
			return true
		}

		if meta.serverName == "" {
			return false
		}

		for _, host := range hosts {
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
