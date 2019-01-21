package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/requestdecorator"
)

func addRoute(ctx context.Context, router *mux.Router, rule string, priority int, handler http.Handler) error {
	matchers, err := parseRule(rule)
	if err != nil {
		return err
	}

	if len(matchers) == 0 {
		return fmt.Errorf("invalid rule: %s", rule)
	}

	if priority == 0 {
		priority = len(rule)
	}

	route := router.NewRoute().Handler(handler).Priority(priority)
	for _, matcher := range matchers {
		matcher(route)
		if route.GetError() != nil {
			log.FromContext(ctx).Error(route.GetError())
		}
	}

	return nil
}

func parseRule(rule string) ([]func(*mux.Route), error) {
	funcs := map[string]func(*mux.Route, ...string){
		"Host":          host,
		"HostRegexp":    hostRegexp,
		"Path":          path,
		"PathPrefix":    pathPrefix,
		"Method":        methods,
		"Headers":       headers,
		"HeadersRegexp": headersRegexp,
		"Query":         query,
	}

	splitRule := func(c rune) bool {
		return c == ';'
	}
	parsedRules := strings.FieldsFunc(rule, splitRule)

	var matchers []func(*mux.Route)

	for _, expression := range parsedRules {
		expParts := strings.Split(expression, ":")
		if len(expParts) > 1 && len(expParts[1]) > 0 {
			if fn, ok := funcs[expParts[0]]; ok {

				parseOr := func(c rune) bool {
					return c == ','
				}

				exp := strings.FieldsFunc(strings.Join(expParts[1:], ":"), parseOr)

				var trimmedExp []string
				for _, value := range exp {
					trimmedExp = append(trimmedExp, strings.TrimSpace(value))
				}

				// FIXME struct for onhostrule ?
				matcher := func(rt *mux.Route) {
					fn(rt, trimmedExp...)
				}

				matchers = append(matchers, matcher)
			} else {
				return nil, fmt.Errorf("invalid matcher: %s", expression)
			}
		}
	}

	return matchers, nil
}

func path(route *mux.Route, paths ...string) {
	rt := route.Subrouter()
	for _, path := range paths {
		tmpRt := rt.Path(path)
		if tmpRt.GetError() != nil {
			log.WithoutContext().WithField("paths", strings.Join(paths, ",")).Error(tmpRt.GetError())
		}
	}
}

func pathPrefix(route *mux.Route, paths ...string) {
	rt := route.Subrouter()
	for _, path := range paths {
		tmpRt := rt.PathPrefix(path)
		if tmpRt.GetError() != nil {
			log.WithoutContext().WithField("paths", strings.Join(paths, ",")).Error(tmpRt.GetError())
		}
	}
}

func host(route *mux.Route, hosts ...string) {
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
}

func hostRegexp(route *mux.Route, hosts ...string) {
	router := route.Subrouter()
	for _, host := range hosts {
		router.Host(host)
	}
}

func methods(route *mux.Route, methods ...string) {
	route.Methods(methods...)
}

func headers(route *mux.Route, headers ...string) {
	route.Headers(headers...)
}

func headersRegexp(route *mux.Route, headers ...string) {
	route.HeadersRegexp(headers...)
}

func query(route *mux.Route, query ...string) {
	var queries []string
	for _, elem := range query {
		queries = append(queries, strings.Split(elem, "=")...)
	}

	route.Queries(queries...)
}
