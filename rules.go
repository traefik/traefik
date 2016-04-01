package main

import (
	"errors"
	"github.com/gorilla/mux"
	"reflect"
	"strings"
)

// Rules holds rule parsing and configuration
type Rules struct {
	route *serverRoute
}

func (r *Rules) host(host string) *mux.Route {
	return r.route.route.Host(host)
}

func (r *Rules) path(path string) *mux.Route {
	return r.route.route.Path(path)
}

func (r *Rules) pathPrefix(path string) *mux.Route {
	return r.route.route.PathPrefix(path)
}

func (r *Rules) pathStrip(path string) *mux.Route {
	r.route.stripPrefix = path
	return r.route.route.Path(path)
}

func (r *Rules) pathPrefixStrip(path string) *mux.Route {
	r.route.stripPrefix = path
	return r.route.route.PathPrefix(path)
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
		"Path":            r.path,
		"PathStrip":       r.pathStrip,
		"PathPrefix":      r.pathPrefix,
		"PathPrefixStrip": r.pathPrefixStrip,
		"Methods":         r.methods,
		"Headers":         r.headers,
		"HeadersRegexp":   r.headersRegexp,
	}
	f := func(c rune) bool {
		return c == ':' || c == '='
	}
	// get function
	parsedFunctions := strings.FieldsFunc(expression, f)
	if len(parsedFunctions) != 2 {
		return nil, errors.New("Error parsing rule: " + expression)
	}
	parsedFunction, ok := functions[parsedFunctions[0]]
	if !ok {
		return nil, errors.New("Error parsing rule: " + expression + ". Unknow function: " + parsedFunctions[0])
	}

	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	parsedArgs := strings.FieldsFunc(parsedFunctions[1], fargs)
	if len(parsedArgs) == 0 {
		return nil, errors.New("Error parsing args from rule: " + expression)
	}

	inputs := make([]reflect.Value, len(parsedArgs))
	for i := range parsedArgs {
		inputs[i] = reflect.ValueOf(parsedArgs[i])
	}
	method := reflect.ValueOf(parsedFunction)
	if method.IsValid() {
		return method.Call(inputs)[0].Interface().(*mux.Route), nil
	}
	return nil, errors.New("Method not found: " + parsedFunctions[0])
}
