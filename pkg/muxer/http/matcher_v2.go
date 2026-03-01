package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/ip"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
)

var httpFuncsV2 = matcherBuilderFuncs{
	"Host":          hostV2,
	"HostHeader":    hostV2,
	"HostRegexp":    hostRegexpV2,
	"ClientIP":      clientIPV2,
	"Path":          pathV2,
	"PathPrefix":    pathPrefixV2,
	"Method":        methodsV2,
	"Headers":       headersV2,
	"HeadersRegexp": headersRegexpV2,
	"Query":         queryV2,
}

func pathV2(tree *matchersTree, paths ...string) error {
	var routes []*mux.Route

	for _, path := range paths {
		route := mux.NewRouter().UseRoutingPath().NewRoute()

		if err := route.Path(path).GetError(); err != nil {
			return err
		}

		routes = append(routes, route)
	}

	tree.matcher = func(req *http.Request) bool {
		for _, route := range routes {
			if route.Match(req, &mux.RouteMatch{}) {
				return true
			}
		}

		return false
	}

	return nil
}

func pathPrefixV2(tree *matchersTree, paths ...string) error {
	var routes []*mux.Route

	for _, path := range paths {
		route := mux.NewRouter().UseRoutingPath().NewRoute()

		if err := route.PathPrefix(path).GetError(); err != nil {
			return err
		}

		routes = append(routes, route)
	}

	tree.matcher = func(req *http.Request) bool {
		for _, route := range routes {
			if route.Match(req, &mux.RouteMatch{}) {
				return true
			}
		}

		return false
	}

	return nil
}

func hostV2(tree *matchersTree, hosts ...string) error {
	for i, host := range hosts {
		if !IsASCII(host) {
			return fmt.Errorf("invalid value %q for \"Host\" matcher, non-ASCII characters are not allowed", host)
		}

		hosts[i] = strings.ToLower(host)
	}

	tree.matcher = func(req *http.Request) bool {
		reqHost := requestdecorator.GetCanonizedHost(req.Context())
		if len(reqHost) == 0 {
			// If the request is an HTTP/1.0 request, then a Host may not be defined.
			if req.ProtoAtLeast(1, 1) {
				log.Ctx(req.Context()).Warn().Msgf("Could not retrieve CanonizedHost, rejecting %s", req.Host)
			}

			return false
		}

		flatH := requestdecorator.GetCNAMEFlatten(req.Context())
		if len(flatH) > 0 {
			for _, host := range hosts {
				if strings.EqualFold(reqHost, host) || strings.EqualFold(flatH, host) {
					return true
				}
				log.Ctx(req.Context()).Debug().Msgf("CNAMEFlattening: request %s which resolved to %s, is not matched to route %s", reqHost, flatH, host)
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
	}

	return nil
}

func clientIPV2(tree *matchersTree, clientIPs ...string) error {
	checker, err := ip.NewChecker(clientIPs)
	if err != nil {
		return fmt.Errorf("could not initialize IP Checker for \"ClientIP\" matcher: %w", err)
	}

	strategy := ip.RemoteAddrStrategy{}

	tree.matcher = func(req *http.Request) bool {
		ok, err := checker.Contains(strategy.GetIP(req))
		if err != nil {
			log.Ctx(req.Context()).Warn().Err(err).Msg("\"ClientIP\" matcher: could not match remote address")
			return false
		}

		return ok
	}

	return nil
}

func methodsV2(tree *matchersTree, methods ...string) error {
	route := mux.NewRouter().NewRoute()
	route.Methods(methods...)
	if err := route.GetError(); err != nil {
		return err
	}

	tree.matcher = func(req *http.Request) bool {
		return route.Match(req, &mux.RouteMatch{})
	}

	return nil
}

func headersV2(tree *matchersTree, headers ...string) error {
	route := mux.NewRouter().NewRoute()
	route.Headers(headers...)
	if err := route.GetError(); err != nil {
		return err
	}

	tree.matcher = func(req *http.Request) bool {
		return route.Match(req, &mux.RouteMatch{})
	}

	return nil
}

func queryV2(tree *matchersTree, query ...string) error {
	var queries []string
	for _, elem := range query {
		queries = append(queries, strings.SplitN(elem, "=", 2)...)
	}

	route := mux.NewRouter().NewRoute()
	route.Queries(queries...)
	if err := route.GetError(); err != nil {
		return err
	}

	tree.matcher = func(req *http.Request) bool {
		return route.Match(req, &mux.RouteMatch{})
	}

	return nil
}

func hostRegexpV2(tree *matchersTree, hosts ...string) error {
	router := mux.NewRouter()

	for _, host := range hosts {
		if !IsASCII(host) {
			return fmt.Errorf("invalid value %q for HostRegexp matcher, non-ASCII characters are not allowed", host)
		}

		tmpRt := router.Host(host)
		if tmpRt.GetError() != nil {
			return tmpRt.GetError()
		}
	}

	tree.matcher = func(req *http.Request) bool {
		return router.Match(req, &mux.RouteMatch{})
	}

	return nil
}

func headersRegexpV2(tree *matchersTree, headers ...string) error {
	route := mux.NewRouter().NewRoute()
	route.HeadersRegexp(headers...)
	if err := route.GetError(); err != nil {
		return err
	}

	tree.matcher = func(req *http.Request) bool {
		return route.Match(req, &mux.RouteMatch{})
	}

	return nil
}
