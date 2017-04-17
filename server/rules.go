package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/mux"
	"github.com/containous/traefik/types"
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
			if types.CanonicalDomain(reqHost) == types.CanonicalDomain(host) {
				return true
			}
		}
		return false
	})
}

func (r *Rules) hostRegexp(hosts ...string) *mux.Route {
	router := r.route.route.Subrouter()
	for _, host := range hosts {
		router.Host(types.CanonicalDomain(host))
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

func (r *Rules) addPrefix(paths ...string) *mux.Route {
	for _, path := range paths {
		r.route.addPrefix = path
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

func (r *Rules) parseRules(expression string, onRule func(functionName string, function interface{}, arguments []string) error) error {
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
		"AddPrefix":       r.addPrefix,
	}

	if len(expression) == 0 {
		return errors.New("Empty rule")
	}

	f := func(c rune) bool {
		return c == ':'
	}

	// Allow multiple rules separated by ;
	splitRule := func(c rune) bool {
		return c == ';'
	}

	parsedRules := strings.FieldsFunc(expression, splitRule)

	for _, rule := range parsedRules {
		// get function
		parsedFunctions := strings.FieldsFunc(rule, f)
		if len(parsedFunctions) == 0 {
			return errors.New("Error parsing rule: '" + rule + "'")
		}
		functionName := strings.TrimSpace(parsedFunctions[0])
		parsedFunction, ok := functions[functionName]
		if !ok {
			return errors.New("Error parsing rule: '" + rule + "'. Unknown function: '" + parsedFunctions[0] + "'")
		}
		parsedFunctions = append(parsedFunctions[:0], parsedFunctions[1:]...)
		fargs := func(c rune) bool {
			return c == ','
		}
		// get function
		parsedArgs := strings.FieldsFunc(strings.Join(parsedFunctions, ":"), fargs)
		if len(parsedArgs) == 0 {
			return errors.New("Error parsing args from rule: '" + rule + "'")
		}

		for i := range parsedArgs {
			parsedArgs[i] = strings.TrimSpace(parsedArgs[i])
		}

		err := onRule(functionName, parsedFunction, parsedArgs)
		if err != nil {
			return fmt.Errorf("Parsing error on rule: %v", err)
		}
	}
	return nil
}

// Parse parses rules expressions
func (r *Rules) Parse(expression string) (*mux.Route, error) {
	var resultRoute *mux.Route
	err := r.parseRules(expression, func(functionName string, function interface{}, arguments []string) error {
		inputs := make([]reflect.Value, len(arguments))
		for i := range arguments {
			inputs[i] = reflect.ValueOf(arguments[i])
		}
		method := reflect.ValueOf(function)
		if method.IsValid() {
			resultRoute = method.Call(inputs)[0].Interface().(*mux.Route)
			if r.err != nil {
				return r.err
			}
			if resultRoute.GetError() != nil {
				return resultRoute.GetError()
			}

		} else {
			return errors.New("Method not found: '" + functionName + "'")
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error parsing rule: %v", err)
	}
	return resultRoute, nil
}

// ParseDomains parses rules expressions and returns domains
func (r *Rules) ParseDomains(expression string) ([]string, error) {
	domains := []string{}
	err := r.parseRules(expression, func(functionName string, function interface{}, arguments []string) error {
		if functionName == "Host" {
			domains = append(domains, arguments...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error parsing domains: %v", err)
	}
	return fun.Map(types.CanonicalDomain, domains).([]string), nil
}
