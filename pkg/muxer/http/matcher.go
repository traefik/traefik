package http

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/ip"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
)

var httpFuncs = matcherBuilderFuncs{
	"ClientIP":     expectNParameters(clientIP, 1),
	"Method":       expectNParameters(method, 1),
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

func expectNParameters(fn func(*matchersTree, ...string) error, n ...int) func(*matchersTree, ...string) error {
	return func(tree *matchersTree, s ...string) error {
		if !slices.Contains(n, len(s)) {
			return fmt.Errorf("unexpected number of parameters; got %d, expected one of %v", len(s), n)
		}

		return fn(tree, s...)
	}
}

func clientIP(tree *matchersTree, clientIP ...string) error {
	checker, err := ip.NewChecker(clientIP)
	if err != nil {
		return fmt.Errorf("initializing IP checker for ClientIP matcher: %w", err)
	}

	strategy := ip.RemoteAddrStrategy{}

	tree.matcher = func(req *http.Request) bool {
		ok, err := checker.Contains(strategy.GetIP(req))
		if err != nil {
			log.Ctx(req.Context()).Warn().Err(err).Msg("ClientIP matcher: could not match remote address")
			return false
		}

		return ok
	}

	return nil
}

func method(tree *matchersTree, methods ...string) error {
	method := strings.ToUpper(methods[0])

	tree.matcher = func(req *http.Request) bool {
		return method == req.Method
	}

	return nil
}

func host(tree *matchersTree, hosts ...string) error {
	host := hosts[0]

	if !IsASCII(host) {
		return fmt.Errorf("invalid value %q for Host matcher, non-ASCII characters are not allowed", host)
	}

	host = strings.ToLower(host)

	tree.matcher = func(req *http.Request) bool {
		reqHost := requestdecorator.GetCanonizedHost(req.Context())
		if len(reqHost) == 0 {
			return false
		}

		if reqHost == host {
			return true
		}

		flatH := requestdecorator.GetCNAMEFlatten(req.Context())
		if len(flatH) > 0 {
			return strings.EqualFold(flatH, host)
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
	}

	return nil
}

func hostRegexp(tree *matchersTree, hosts ...string) error {
	host := hosts[0]

	if !IsASCII(host) {
		return fmt.Errorf("invalid value %q for HostRegexp matcher, non-ASCII characters are not allowed", host)
	}

	re, err := regexp.Compile(host)
	if err != nil {
		return fmt.Errorf("compiling HostRegexp matcher: %w", err)
	}

	tree.matcher = func(req *http.Request) bool {
		return re.MatchString(requestdecorator.GetCanonizedHost(req.Context())) ||
			re.MatchString(requestdecorator.GetCNAMEFlatten(req.Context()))
	}

	return nil
}

func path(tree *matchersTree, paths ...string) error {
	path := paths[0]

	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %q does not start with a '/'", path)
	}

	tree.matcher = func(req *http.Request) bool {
		routingPath := getRoutingPath(req)
		return routingPath != nil && *routingPath == path
	}

	return nil
}

func pathRegexp(tree *matchersTree, paths ...string) error {
	path := paths[0]

	re, err := regexp.Compile(path)
	if err != nil {
		return fmt.Errorf("compiling PathPrefix matcher: %w", err)
	}

	tree.matcher = func(req *http.Request) bool {
		routingPath := getRoutingPath(req)
		return routingPath != nil && re.MatchString(*routingPath)
	}

	return nil
}

func pathPrefix(tree *matchersTree, paths ...string) error {
	path := paths[0]

	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %q does not start with a '/'", path)
	}

	tree.matcher = func(req *http.Request) bool {
		routingPath := getRoutingPath(req)
		return routingPath != nil && strings.HasPrefix(*routingPath, path)
	}

	return nil
}

func header(tree *matchersTree, headers ...string) error {
	key, value := http.CanonicalHeaderKey(headers[0]), headers[1]

	tree.matcher = func(req *http.Request) bool {
		for _, headerValue := range req.Header[key] {
			if headerValue == value {
				return true
			}
		}

		return false
	}

	return nil
}

func headerRegexp(tree *matchersTree, headers ...string) error {
	key, value := http.CanonicalHeaderKey(headers[0]), headers[1]

	re, err := regexp.Compile(value)
	if err != nil {
		return fmt.Errorf("compiling HeaderRegexp matcher: %w", err)
	}

	tree.matcher = func(req *http.Request) bool {
		for _, headerValue := range req.Header[key] {
			if re.MatchString(headerValue) {
				return true
			}
		}

		return false
	}

	return nil
}

func query(tree *matchersTree, queries ...string) error {
	key := queries[0]

	var value string
	if len(queries) == 2 {
		value = queries[1]
	}

	tree.matcher = func(req *http.Request) bool {
		values, ok := req.URL.Query()[key]
		if !ok {
			return false
		}

		return slices.Contains(values, value)
	}

	return nil
}

func queryRegexp(tree *matchersTree, queries ...string) error {
	if len(queries) == 1 {
		return query(tree, queries...)
	}

	key, value := queries[0], queries[1]

	re, err := regexp.Compile(value)
	if err != nil {
		return fmt.Errorf("compiling QueryRegexp matcher: %w", err)
	}

	tree.matcher = func(req *http.Request) bool {
		values, ok := req.URL.Query()[key]
		if !ok {
			return false
		}

		idx := slices.IndexFunc(values, func(value string) bool {
			return re.MatchString(value)
		})

		return idx >= 0
	}

	return nil
}

// IsASCII checks if the given string contains only ASCII characters.
func IsASCII(s string) bool {
	for i := range len(s) {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}

	return true
}
