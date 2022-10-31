package http

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares/requestdecorator"
	"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/vulcand/predicate"
)

const hostMatcher = "Host"

var httpFuncs = map[string]func(*mux.Route, ...string) error{
	hostMatcher:     host,
	"HostHeader":    host,
	"HostRegexp":    hostRegexp,
	"ClientIP":      clientIP,
	"Path":          path,
	"PathPrefix":    pathPrefix,
	"Method":        methods,
	"Headers":       headers,
	"HeadersRegexp": headersRegexp,
	"Query":         query,
}

// Muxer handles routing with rules.
type Muxer struct {
	*mux.Router
	parser predicate.Parser
}

// NewMuxer returns a new muxer instance.
func NewMuxer() (*Muxer, error) {
	var matchers []string
	for matcher := range httpFuncs {
		matchers = append(matchers, matcher)
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		return nil, err
	}

	return &Muxer{
		Router: mux.NewRouter().SkipClean(true),
		parser: parser,
	}, nil
}

// AddRoute add a new route to the router.
func (r *Muxer) AddRoute(rule string, priority int, handler http.Handler) error {
	parse, err := r.parser.Parse(rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return fmt.Errorf("error while parsing rule %s", rule)
	}

	if priority == 0 {
		priority = len(rule)
	}

	route := r.NewRoute().Handler(handler).Priority(priority)

	err = addRuleOnRoute(route, buildTree())
	if err != nil {
		route.BuildOnly()
		return err
	}

	return nil
}

// ParseDomains extract domains from rule.
func ParseDomains(rule string) ([]string, error) {
	var matchers []string
	for matcher := range httpFuncs {
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
		return nil, fmt.Errorf("error while parsing rule %s", rule)
	}

	return buildTree().ParseMatchers([]string{hostMatcher}), nil
}

func path(route *mux.Route, paths ...string) error {
	rt := route.Subrouter()

	for _, path := range paths {
		if err := rt.Path(path).GetError(); err != nil {
			return err
		}
	}

	return nil
}

func pathPrefix(route *mux.Route, paths ...string) error {
	rt := route.Subrouter()

	for _, path := range paths {
		if err := rt.PathPrefix(path).GetError(); err != nil {
			return err
		}
	}

	return nil
}

func host(route *mux.Route, hosts ...string) error {
	for i, host := range hosts {
		if !IsASCII(host) {
			return fmt.Errorf("invalid value %q for \"Host\" matcher, non-ASCII characters are not allowed", host)
		}

		hosts[i] = strings.ToLower(host)
	}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		reqHost := requestdecorator.GetCanonizedHost(req.Context())
		if len(reqHost) == 0 {
			// If the request is an HTTP/1.0 request, then a Host may not be defined.
			if req.ProtoAtLeast(1, 1) {
				log.FromContext(req.Context()).Warnf("Could not retrieve CanonizedHost, rejecting %s", req.Host)
			}

			return false
		}

		flatH := requestdecorator.GetCNAMEFlatten(req.Context())
		if len(flatH) > 0 {
			for _, host := range hosts {
				if strings.EqualFold(reqHost, host) || strings.EqualFold(flatH, host) {
					return true
				}
				log.FromContext(req.Context()).Debugf("CNAMEFlattening: request %s which resolved to %s, is not matched to route %s", reqHost, flatH, host)
			}
			return false
		}

		for _, host := range hosts {
			if reqHost == host {
				return true
			}

			// Check for match on trailing period on host
			if last := len(host) - 1; last >= 0 && host[last] == '.' {
				h := host[:last]
				if reqHost == h {
					return true
				}
			}

			// Check for match on trailing period on request
			if last := len(reqHost) - 1; last >= 0 && reqHost[last] == '.' {
				h := reqHost[:last]
				if h == host {
					return true
				}
			}
		}
		return false
	})
	return nil
}

func clientIP(route *mux.Route, clientIPs ...string) error {
	checker, err := ip.NewChecker(clientIPs)
	if err != nil {
		return fmt.Errorf("could not initialize IP Checker for \"ClientIP\" matcher: %w", err)
	}

	strategy := ip.RemoteAddrStrategy{}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		ok, err := checker.Contains(strategy.GetIP(req))
		if err != nil {
			log.FromContext(req.Context()).Warnf("\"ClientIP\" matcher: could not match remote address : %w", err)
			return false
		}

		return ok
	})

	return nil
}

func hostRegexp(route *mux.Route, hosts ...string) error {
	router := route.Subrouter()
	for _, host := range hosts {
		if !IsASCII(host) {
			return fmt.Errorf("invalid value %q for HostRegexp matcher, non-ASCII characters are not allowed", host)
		}

		tmpRt := router.Host(host)
		if tmpRt.GetError() != nil {
			return tmpRt.GetError()
		}
	}
	return nil
}

func methods(route *mux.Route, methods ...string) error {
	return route.Methods(methods...).GetError()
}

func headers(route *mux.Route, headers ...string) error {
	return route.Headers(headers...).GetError()
}

func headersRegexp(route *mux.Route, headers ...string) error {
	return route.HeadersRegexp(headers...).GetError()
}

func query(route *mux.Route, query ...string) error {
	var queries []string
	for _, elem := range query {
		queries = append(queries, strings.SplitN(elem, "=", 2)...)
	}

	route.Queries(queries...)
	// Queries can return nil so we can't chain the GetError()
	return route.GetError()
}

func addRuleOnRouter(router *mux.Router, rule *rules.Tree) error {
	switch rule.Matcher {
	case "and":
		route := router.NewRoute()
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
			return not(httpFuncs[rule.Matcher])(router.NewRoute(), rule.Value...)
		}

		return httpFuncs[rule.Matcher](router.NewRoute(), rule.Value...)
	}
}

func not(m func(*mux.Route, ...string) error) func(*mux.Route, ...string) error {
	return func(r *mux.Route, v ...string) error {
		router := mux.NewRouter()
		err := m(router.NewRoute(), v...)
		if err != nil {
			return err
		}
		r.MatcherFunc(func(req *http.Request, ma *mux.RouteMatch) bool {
			return !router.Match(req, ma)
		})
		return nil
	}
}

func addRuleOnRoute(route *mux.Route, rule *rules.Tree) error {
	switch rule.Matcher {
	case "and":
		err := addRuleOnRoute(route, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRoute(route, rule.RuleRight)
	case "or":
		subRouter := route.Subrouter()

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
			return not(httpFuncs[rule.Matcher])(route, rule.Value...)
		}

		return httpFuncs[rule.Matcher](route, rule.Value...)
	}
}

// IsASCII checks if the given string contains only ASCII characters.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}

	return true
}
