package rules

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

// Rules holds rule parsing and configuration
type Rules struct {
	Route *types.ServerRoute
	err   error
}

func (r *Rules) host(hosts ...string) *mux.Route {
	return r.Route.Route.MatcherFunc(func(req *http.Request, route *mux.RouteMatch) bool {
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

const (
	maxUint             = ^uint64(0)
	hashRangeHeaderType = "header"
)

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func getMinMax(s string) (min, max int, err error) {
	matchParts := strings.Split(s, "-")
	if len(matchParts) != 2 {
		return 0, 0, fmt.Errorf("failed to parse int range: %v", s)
	}
	min, err = strconv.Atoi(matchParts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse int range: %v", s)
	}
	max, err = strconv.Atoi(matchParts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse int range: %v", s)
	}
	return
}

func (r *Rules) hashedRange(rule ...string) *mux.Route {
	falseMatch := func(req *http.Request, route *mux.RouteMatch) bool { return false }
	args := map[string]string{}
	items := strings.Split(rule[0], " ")
	//expect 4 parts: type, value, match and range
	if len(items) != 4 {
		log.Errorf("Failed parsing hashedRange matcher rule: %v", rule)
		return r.Route.Route.MatcherFunc(falseMatch)
	}
	for _, item := range items {
		parts := strings.Split(item, ":")
		//expect format 'type:header'
		if len(parts) != 2 {
			log.Error("Failed parsing hashedRange matcher rule: %v", rule)
			r.Route.Route.MatcherFunc(falseMatch)
		}
		args[parts[0]] = parts[1]
	}

	//check parts included
	typeMatch, exists := args["type"]
	if !exists {
		log.Errorf("Failed parsing hashedRange matcher rule: %v", rule)
		r.Route.Route.MatcherFunc(falseMatch)
	}
	typeValue, exists := args["value"]
	if !exists {
		log.Errorf("Failed parsing hashedRange matcher rule: %v", rule)
		r.Route.Route.MatcherFunc(falseMatch)
	}
	matchString, exists := args["match"]
	if !exists {
		log.Errorf("Failed parsing hashedRange matcher rule: %v", rule)
		r.Route.Route.MatcherFunc(falseMatch)
	}
	matchLow, matchHigh, err := getMinMax(matchString)
	if err != nil {
		log.Errorf("Failed parsing hashedRange matcher rule: %v", rule)
		r.Route.Route.MatcherFunc(falseMatch)
	}

	rangeString, exists := args["range"]
	if !exists {
		log.Errorf("Failed parsing hashedRange matcher rule: %v", rule)
		r.Route.Route.MatcherFunc(falseMatch)
	}
	rangeLow, rangeHigh, err := getMinMax(rangeString)
	if err != nil {
		log.Errorf("Failed parsing hashedRange matcher rule: %v", rule)
		r.Route.Route.MatcherFunc(falseMatch)
	}

	return r.Route.Route.MatcherFunc(func(req *http.Request, route *mux.RouteMatch) bool {
		var value string
		if typeMatch == hashRangeHeaderType {
			value = req.Header.Get(typeValue)
		}

		hash := hash(value)

		factor := float64(rangeHigh-rangeLow) / float64(maxUint)
		log.Error(factor)
		adjustedResult := (float64(hash) * factor) + float64(rangeLow)

		log.Error(adjustedResult)

		if adjustedResult >= float64(matchLow) && adjustedResult < float64(matchHigh) {
			return true
		}

		return false
	})
}

func (r *Rules) hostRegexp(hosts ...string) *mux.Route {
	router := r.Route.Route.Subrouter()
	for _, host := range hosts {
		router.Host(types.CanonicalDomain(host))
	}
	return r.Route.Route
}

func (r *Rules) path(paths ...string) *mux.Route {
	router := r.Route.Route.Subrouter()
	for _, path := range paths {
		router.Path(strings.TrimSpace(path))
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
	cleanPath := strings.TrimSpace(path)
	// {} are used to define a regex pattern in http://www.gorillatoolkit.org/pkg/mux.
	// if we find a { in the path, that means we use regex, then the gorilla/mux implementation is chosen
	// otherwise, we use a lightweight implementation
	if strings.Contains(cleanPath, "{") {
		router.PathPrefix(cleanPath)
	} else {
		m := &prefixMatcher{prefix: cleanPath}
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
		router.Path(strings.TrimSpace(path))
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
		router.PathPrefix(strings.TrimSpace(path))
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
		"HashedRange":          r.hashedRange,
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

	err := r.parseRules(expression, func(functionName string, function interface{}, arguments []string) error {
		if functionName == "Host" {
			domains = append(domains, arguments...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing domains: %v", err)
	}

	return fun.Map(types.CanonicalDomain, domains).([]string), nil
}
