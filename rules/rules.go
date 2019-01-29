package rules

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/vulcand/predicate"

	"github.com/pkg/errors"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/requestdecorator"
)

var funcs = map[string]func(*mux.Route, ...string) error{
	"Host":          host,
	"HostRegexp":    hostRegexp,
	"Path":          path,
	"PathPrefix":    pathPrefix,
	"Method":        methods,
	"Headers":       headers,
	"HeadersRegexp": headersRegexp,
	"Query":         query,
}

// Router handle routing with rules
type Router struct {
	*mux.Router
	parser predicate.Parser
}

// NewRouter returns a new router instance.
func NewRouter() (*Router, error) {
	parser, err := newParser()
	if err != nil {
		return nil, err
	}

	return &Router{
		Router: mux.NewRouter().SkipClean(true),
		parser: parser,
	}, nil
}

// AddRoute add a new route to the router.
func (r *Router) AddRoute(ctx context.Context, rule string, priority int, handler http.Handler) error {
	parse, err := r.parser.Parse(rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %v", rule, err)
	}

	buildTree, ok := parse.(treeBuilder)
	if !ok {
		return errors.New("cannot parse")
	}

	if priority == 0 {
		priority = len(rule)
	}

	route := r.NewRoute().Handler(handler).Priority(priority)
	return addRuleOnRoute(route, buildTree())
}

type tree struct {
	matcher string
	value   []string
	ruleA   *tree
	ruleB   *tree
}

func path(route *mux.Route, paths ...string) error {
	rt := route.Subrouter()

	for _, path := range paths {
		tmpRt := rt.Path(path)
		if tmpRt.GetError() != nil {
			return tmpRt.GetError()
		}
	}
	return nil
}

func pathPrefix(route *mux.Route, paths ...string) error {
	rt := route.Subrouter()

	for _, path := range paths {
		tmpRt := rt.PathPrefix(path)
		if tmpRt.GetError() != nil {
			return tmpRt.GetError()
		}
	}
	return nil
}

func host(route *mux.Route, hosts ...string) error {
	for i, host := range hosts {
		hosts[i] = strings.ToLower(host)
	}

	route.MatcherFunc(func(req *http.Request, route *mux.RouteMatch) bool {
		reqHost := requestdecorator.GetCanonizedHost(req.Context())
		if len(reqHost) == 0 {
			log.FromContext(req.Context()).Warnf("Could not retrieve CanonizedHost, rejecting %s", req.Host)
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
		}
		return false
	})
	return nil
}

func hostRegexp(route *mux.Route, hosts ...string) error {
	router := route.Subrouter()
	for _, host := range hosts {
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
		queries = append(queries, strings.Split(elem, "=")...)
	}

	route.Queries(queries...)
	// Queries can return nil so we can't chain the GetError()
	return route.GetError()
}

func addRuleOnRouter(router *mux.Router, rule *tree) error {
	switch rule.matcher {
	case "and":
		route := router.NewRoute()
		err := addRuleOnRoute(route, rule.ruleA)
		if err != nil {
			return err
		}

		err = addRuleOnRoute(route, rule.ruleB)
		if err != nil {
			return err
		}
	case "or":
		err := addRuleOnRouter(router, rule.ruleA)
		if err != nil {
			return err
		}

		err = addRuleOnRouter(router, rule.ruleB)
		if err != nil {
			return err
		}
	default:
		err := checkRule(rule)
		if err != nil {
			return err
		}

		return funcs[rule.matcher](router.NewRoute(), rule.value...)
	}

	return nil
}

func addRuleOnRoute(route *mux.Route, rule *tree) error {
	switch rule.matcher {
	case "and":
		err := addRuleOnRoute(route, rule.ruleA)
		if err != nil {
			return err
		}

		err = addRuleOnRoute(route, rule.ruleB)
		if err != nil {
			return err
		}
	case "or":
		subrouter := route.Subrouter()

		err := addRuleOnRouter(subrouter, rule.ruleA)
		if err != nil {
			return err
		}

		err = addRuleOnRouter(subrouter, rule.ruleB)
		if err != nil {
			return err
		}
	default:
		err := checkRule(rule)
		if err != nil {
			return err
		}

		return funcs[rule.matcher](route, rule.value...)
	}

	return nil
}

func checkRule(rule *tree) error {
	if len(rule.value) == 0 {
		return fmt.Errorf("no args for matcher %s", rule.matcher)
	}

	for _, v := range rule.value {
		if len(v) == 0 {
			return fmt.Errorf("empty args for matcher %s, %v", rule.matcher, rule.value)
		}
	}
	return nil
}
