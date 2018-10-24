package rules

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/containous/mux"
	"github.com/containous/traefik/hostresolver"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/types"
)

// Rules holds rule parsing and configuration
type Rules struct {
	Route        *types.ServerRoute
	err          error
	HostResolver *hostresolver.Resolver
}

func (r *Rules) host(hosts ...string) *mux.Route {
	for i, host := range hosts {
		hosts[i] = strings.ToLower(host)
	}

	return r.Route.Route.MatcherFunc(func(req *http.Request, route *mux.RouteMatch) bool {
		reqHost := middlewares.GetCanonizedHost(req.Context())
		if len(reqHost) == 0 {
			return false
		}

		if r.HostResolver != nil && r.HostResolver.CnameFlattening {
			reqH, flatH := r.HostResolver.CNAMEFlatten(reqHost)
			for _, host := range hosts {
				if strings.EqualFold(reqH, host) || strings.EqualFold(flatH, host) {
					return true
				}
				log.Debugf("CNAMEFlattening: request %s which resolved to %s, is not matched to route %s", reqH, flatH, host)
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

func (r *Rules) hostRegexp(hostPatterns ...string) *mux.Route {
	router := r.Route.Route.Subrouter()
	for _, hostPattern := range hostPatterns {
		router.Host(hostPattern)
	}
	return r.Route.Route
}

func (r *Rules) path(paths ...string) *mux.Route {
	router := r.Route.Route.Subrouter()
	for _, path := range paths {
		router.Path(path)
	}
	return r.Route.Route
}

func (r *Rules) pathPrefix(paths ...string) *mux.Route {
	router := r.Route.Route.Subrouter()
	for _, path := range paths {
		buildPath(path, router)
	}
	return r.Route.Route
}

func buildPath(path string, router *mux.Router) {
	// {} are used to define a regex pattern in http://www.gorillatoolkit.org/pkg/mux.
	// if we find a { in the path, that means we use regex, then the gorilla/mux implementation is chosen
	// otherwise, we use a lightweight implementation
	if strings.Contains(path, "{") {
		router.PathPrefix(path)
	} else {
		m := &prefixMatcher{prefix: path}
		router.NewRoute().MatcherFunc(m.Match)
	}
}

type prefixMatcher struct {
	prefix string
}

func (m *prefixMatcher) Match(r *http.Request, _ *mux.RouteMatch) bool {
	return strings.HasPrefix(r.URL.Path, m.prefix) || strings.HasPrefix(r.URL.Path, m.prefix+"/")
}

type bySize []string

func (a bySize) Len() int           { return len(a) }
func (a bySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bySize) Less(i, j int) bool { return len(a[i]) > len(a[j]) }

func (r *Rules) pathStrip(paths ...string) *mux.Route {
	sort.Sort(bySize(paths))
	r.Route.StripPrefixes = paths
	router := r.Route.Route.Subrouter()
	for _, path := range paths {
		router.Path(strings.TrimSpace(path))
	}
	return r.Route.Route
}

func (r *Rules) pathStripRegex(paths ...string) *mux.Route {
	sort.Sort(bySize(paths))
	r.Route.StripPrefixesRegex = paths
	router := r.Route.Route.Subrouter()
	for _, path := range paths {
		router.Path(path)
	}
	return r.Route.Route
}

func (r *Rules) replacePath(paths ...string) *mux.Route {
	for _, path := range paths {
		r.Route.ReplacePath = path
	}
	return r.Route.Route
}

func (r *Rules) replacePathRegex(paths ...string) *mux.Route {
	for _, path := range paths {
		r.Route.ReplacePathRegex = path
	}
	return r.Route.Route
}

func (r *Rules) addPrefix(paths ...string) *mux.Route {
	for _, path := range paths {
		r.Route.AddPrefix = path
	}
	return r.Route.Route
}

func (r *Rules) pathPrefixStrip(paths ...string) *mux.Route {
	sort.Sort(bySize(paths))
	r.Route.StripPrefixes = paths
	router := r.Route.Route.Subrouter()
	for _, path := range paths {
		buildPath(path, router)
	}
	return r.Route.Route
}

func (r *Rules) pathPrefixStripRegex(paths ...string) *mux.Route {
	sort.Sort(bySize(paths))
	r.Route.StripPrefixesRegex = paths
	router := r.Route.Route.Subrouter()
	for _, path := range paths {
		router.PathPrefix(path)
	}
	return r.Route.Route
}

func (r *Rules) methods(methods ...string) *mux.Route {
	return r.Route.Route.Methods(methods...)
}

func (r *Rules) headers(headers ...string) *mux.Route {
	return r.Route.Route.Headers(headers...)
}

func (r *Rules) headersRegexp(headers ...string) *mux.Route {
	return r.Route.Route.HeadersRegexp(headers...)
}

func (r *Rules) query(query ...string) *mux.Route {
	var queries []string
	for _, elem := range query {
		queries = append(queries, strings.Split(elem, "=")...)
	}

	return r.Route.Route.Queries(queries...)
}

func (r *Rules) parseRules(expression string, onRule func(functionName string, function interface{}, arguments []string) error) error {
	functions := map[string]interface{}{
		"Host":                 r.host,
		"HostRegexp":           r.hostRegexp,
		"Path":                 r.path,
		"PathStrip":            r.pathStrip,
		"PathStripRegex":       r.pathStripRegex,
		"PathPrefix":           r.pathPrefix,
		"PathPrefixStrip":      r.pathPrefixStrip,
		"PathPrefixStripRegex": r.pathPrefixStripRegex,
		"Method":               r.methods,
		"Headers":              r.headers,
		"HeadersRegexp":        r.headersRegexp,
		"AddPrefix":            r.addPrefix,
		"ReplacePath":          r.replacePath,
		"ReplacePathRegex":     r.replacePathRegex,
		"Query":                r.query,
	}

	if len(expression) == 0 {
		return errors.New("empty rule")
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
			return fmt.Errorf("error parsing rule: '%s'", rule)
		}

		functionName := strings.TrimSpace(parsedFunctions[0])
		parsedFunction, ok := functions[functionName]
		if !ok {
			return fmt.Errorf("error parsing rule: '%s'. Unknown function: '%s'", rule, parsedFunctions[0])
		}
		parsedFunctions = append(parsedFunctions[:0], parsedFunctions[1:]...)

		// get function
		fargs := func(c rune) bool {
			return c == ','
		}
		parsedArgs := strings.FieldsFunc(strings.Join(parsedFunctions, ":"), fargs)
		if len(parsedArgs) == 0 {
			return fmt.Errorf("error parsing args from rule: '%s'", rule)
		}

		for i := range parsedArgs {
			parsedArgs[i] = strings.TrimSpace(parsedArgs[i])
		}

		err := onRule(functionName, parsedFunction, parsedArgs)
		if err != nil {
			return fmt.Errorf("parsing error on rule: %v", err)
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
			if resultRoute == nil {
				return fmt.Errorf("invalid expression: %s", expression)
			}
			if resultRoute.GetError() != nil {
				return resultRoute.GetError()
			}
		} else {
			return fmt.Errorf("method not found: '%s'", functionName)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing rule: %v", err)
	}

	return resultRoute, nil
}

// ParseDomains parses rules expressions and returns domains
func (r *Rules) ParseDomains(expression string) ([]string, error) {
	var domains []string
	isHostRule := false

	err := r.parseRules(expression, func(functionName string, function interface{}, arguments []string) error {
		if functionName == "Host" {
			isHostRule = true
			domains = append(domains, arguments...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing domains: %v", err)
	}

	var cleanDomains []string
	for _, domain := range domains {
		canonicalDomain := strings.ToLower(domain)
		if len(canonicalDomain) > 0 {
			cleanDomains = append(cleanDomains, canonicalDomain)
		}
	}

	// Return an error if an Host rule is detected but no domain are parsed
	if isHostRule && len(cleanDomains) == 0 {
		return nil, fmt.Errorf("unable to parse correctly the domains in the Host rule from %q", expression)
	}

	return cleanDomains, nil
}
