package main

import (
	"errors"
	"github.com/containous/mux"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

// Rules holds rule parsing and configuration
type Rules struct {
	route *serverRoute
	err   error
}

func (r *Rules) host(hosts ...string) *mux.Route {
	return r.route.route.MatcherFunc(func(req *http.Request, route *mux.RouteMatch) bool {
		reqHost, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			reqHost = req.Host
		}
		for _, host := range hosts {
			if reqHost == strings.TrimSpace(host) {
				return true
			}
		}
		return false
	})
}

func (r *Rules) hostRegexp(hosts ...string) *mux.Route {
	router := r.route.route.Subrouter()
	for _, host := range hosts {
		router.Host(strings.TrimSpace(host))
	}
	return r.route.route
}

func (r *Rules) path(paths ...string) *mux.Route {
	router := r.route.route.Subrouter()
	for _, path := range paths {
		router.Path(strings.TrimSpace(path))
	}
	return r.route.route
}

func (r *Rules) pathPrefix(paths ...string) *mux.Route {
	router := r.route.route.Subrouter()
	for _, path := range paths {
		router.PathPrefix(strings.TrimSpace(path))
	}
	return r.route.route
}

type bySize []string

func (a bySize) Len() int           { return len(a) }
func (a bySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bySize) Less(i, j int) bool { return len(a[i]) > len(a[j]) }

func (r *Rules) pathStrip(paths ...string) *mux.Route {
	sort.Sort(bySize(paths))
	r.route.stripPrefixes = paths
	router := r.route.route.Subrouter()
	for _, path := range paths {
		router.Path(strings.TrimSpace(path))
	}
	return r.route.route
}

func (r *Rules) pathPrefixStrip(paths ...string) *mux.Route {
	sort.Sort(bySize(paths))
	r.route.stripPrefixes = paths
	router := r.route.route.Subrouter()
	for _, path := range paths {
		router.PathPrefix(strings.TrimSpace(path))
	}
	return r.route.route
}

func (r *Rules) methods(methods ...string) *mux.Route {
	return r.route.route.Methods(methods...)
}

func (r *Rules) headers(headers ...string) *mux.Route {
	return r.route.route.Headers(headers...)
}

func (r *Rules) headersRegexp(headers ...string) *mux.Route {
	return r.route.route.HeadersRegexp(headers...)
}

// Parse parses rules expressions
func (r *Rules) Parse(expression string) (*mux.Route, error) {
	functions := map[string]interface{}{
		"Host":            r.host,
		"HostRegexp":      r.hostRegexp,
		"Path":            r.path,
		"PathStrip":       r.pathStrip,
		"PathPrefix":      r.pathPrefix,
		"PathPrefixStrip": r.pathPrefixStrip,
		"Method":          r.methods,
		"Headers":         r.headers,
		"HeadersRegexp":   r.headersRegexp,
	}
	f := func(c rune) bool {
		return c == ':'
	}

	// Allow multiple rules separated by ;
	splitRule := func(c rune) bool {
		return c == ';'
	}

	parsedRules := strings.FieldsFunc(expression, splitRule)

	var resultRoute *mux.Route

	for _, rule := range parsedRules {
		// get function
		parsedFunctions := strings.FieldsFunc(rule, f)
		if len(parsedFunctions) == 0 {
			return nil, errors.New("Error parsing rule: " + rule)
		}
		parsedFunction, ok := functions[parsedFunctions[0]]
		if !ok {
			return nil, errors.New("Error parsing rule: " + rule + ". Unknown function: " + parsedFunctions[0])
		}
		parsedFunctions = append(parsedFunctions[:0], parsedFunctions[1:]...)
		fargs := func(c rune) bool {
			return c == ','
		}
		// get function
		parsedArgs := strings.FieldsFunc(strings.Join(parsedFunctions, ":"), fargs)
		if len(parsedArgs) == 0 {
			return nil, errors.New("Error parsing args from rule: " + rule)
		}

		inputs := make([]reflect.Value, len(parsedArgs))
		for i := range parsedArgs {
			inputs[i] = reflect.ValueOf(parsedArgs[i])
		}
		method := reflect.ValueOf(parsedFunction)
		if method.IsValid() {
			resultRoute = method.Call(inputs)[0].Interface().(*mux.Route)
			if r.err != nil {
				return nil, r.err
			}
			if resultRoute.GetError() != nil {
				return nil, resultRoute.GetError()
			}

		} else {
			return nil, errors.New("Method not found: " + parsedFunctions[0])
		}
	}
	return resultRoute, nil
}
