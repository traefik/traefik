package http

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/middlewares/requestdecorator"
	"golang.org/x/exp/slices"
)

var httpFuncs = map[string]func(*mux.Route, ...string) error{
	"ClientIP":     expectNParameters(clientIP, 1),
	"Method":       expectNParameters(methods, 1),
	"Host":         expectNParameters(host, 1),
	"HostRegexp":   expectNParameters(hostRegexp, 1),
	"Path":         expectNParameters(path, 1),
	"PathRegexp":   expectNParameters(pathRegexp, 1),
	"PathPrefix":   expectNParameters(pathPrefix, 1),
	"Header":       expectNParameters(header, 2),
	"HeaderRegexp": expectNParameters(headerRegexp, 2),
	"Query":        expectNParameters(query, 1, 2),
	"QueryRegexp":  expectNParameters(queryRegexp, 1, 2),
}

func expectNParameters(fn func(*mux.Route, ...string) error, n ...int) func(*mux.Route, ...string) error {
	return func(route *mux.Route, s ...string) error {
		if !slices.Contains(n, len(s)) {
			return fmt.Errorf("unexpected number of parameters; got %d, expected one of %v", len(s), n)
		}

		return fn(route, s...)
	}
}

func clientIP(route *mux.Route, clientIP ...string) error {
	checker, err := ip.NewChecker(clientIP)
	if err != nil {
		return fmt.Errorf("initializing IP checker for ClientIP matcher: %w", err)
	}

	strategy := ip.RemoteAddrStrategy{}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		ok, err := checker.Contains(strategy.GetIP(req))
		if err != nil {
			log.Ctx(req.Context()).Warn().Err(err).Msg("ClientIP matcher: could not match remote address")
			return false
		}

		return ok
	})

	return nil
}

func methods(route *mux.Route, methods ...string) error {
	return route.Methods(methods...).GetError()
}

func host(route *mux.Route, hosts ...string) error {
	host := hosts[0]

	if !IsASCII(host) {
		return fmt.Errorf("invalid value %q for Host matcher, non-ASCII characters are not allowed", host)
	}

	host = strings.ToLower(host)

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		reqHost := requestdecorator.GetCanonizedHost(req.Context())
		if len(reqHost) == 0 {
			// If the request is an HTTP/1.0 request, then a Host may not be defined.
			if req.ProtoAtLeast(1, 1) {
				log.Ctx(req.Context()).Warn().Str("host", req.Host).Msg("Could not retrieve CanonizedHost, rejecting")
			}

			return false
		}

		flatH := requestdecorator.GetCNAMEFlatten(req.Context())
		if len(flatH) > 0 {
			if strings.EqualFold(reqHost, host) || strings.EqualFold(flatH, host) {
				return true
			}

			log.Ctx(req.Context()).Debug().
				Str("host", reqHost).
				Str("flattenHost", flatH).
				Str("matcher", host).
				Msg("CNAMEFlattening: resolved Host does not match")
			return false
		}

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

		return false
	})

	return nil
}

func hostRegexp(route *mux.Route, hosts ...string) error {
	host := hosts[0]

	if !IsASCII(host) {
		return fmt.Errorf("invalid value %q for HostRegexp matcher, non-ASCII characters are not allowed", host)
	}

	re, err := regexp.Compile(host)
	if err != nil {
		return fmt.Errorf("compiling HostRegexp matcher: %w", err)
	}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		return re.MatchString(req.Host)
	})

	return nil
}

func path(route *mux.Route, paths ...string) error {
	path := paths[0]

	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %q does not start with a '/'", path)
	}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		return req.URL.Path == path
	})

	return nil
}

func pathRegexp(route *mux.Route, paths ...string) error {
	path := paths[0]

	re, err := regexp.Compile(path)
	if err != nil {
		return fmt.Errorf("compiling PathPrefix matcher: %w", err)
	}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		return re.MatchString(req.URL.Path)
	})

	return nil
}

func pathPrefix(route *mux.Route, paths ...string) error {
	path := paths[0]

	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %q does not start with a '/'", path)
	}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		return strings.HasPrefix(req.URL.Path, path)
	})

	return nil
}

func header(route *mux.Route, headers ...string) error {
	return route.Headers(headers...).GetError()
}

func headerRegexp(route *mux.Route, headers ...string) error {
	return route.HeadersRegexp(headers...).GetError()
}

func query(route *mux.Route, queries ...string) error {
	key := queries[0]

	var value string
	if len(queries) == 2 {
		value = queries[1]
	}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		values, ok := req.URL.Query()[key]
		if !ok {
			return false
		}

		return slices.Contains(values, value)
	})

	return nil
}

func queryRegexp(route *mux.Route, queries ...string) error {
	if len(queries) == 1 {
		return query(route, queries...)
	}

	key, value := queries[0], queries[1]

	re, err := regexp.Compile(value)
	if err != nil {
		return fmt.Errorf("compiling QueryRegexp matcher: %w", err)
	}

	route.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool {
		values, ok := req.URL.Query()[key]
		if !ok {
			return false
		}

		idx := slices.IndexFunc(values, func(value string) bool {
			return re.MatchString(value)
		})

		return idx >= 0
	})

	return nil
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
