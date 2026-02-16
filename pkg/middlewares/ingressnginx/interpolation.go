package ingressnginx

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// This the list of supported NGINX variables for interpolation.
// It is not exhaustive, but covers the most commonly used ones in Ingress NGINX annotations.
const (
	scheme        = "$scheme"
	host          = "$host"
	httpHeaders   = "$http_"
	hostname      = "$hostname"
	requestURI    = "$request_uri"
	queryString   = "$query_string"
	args          = "$args"
	arg           = "$arg_"
	remoteAddress = "$remote_addr"

	// Variables set by ingress-nginx template.
	bestHTTPHost      = "$best_http_host"
	escapedRequestURI = "$escaped_request_uri"
)

// varRegexp is a regular expression to match NGINX variables in the form of $variable or $variable_name.
var varRegexp = regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`)

// ReplaceVariables replaces NGINX variables in the given string with their corresponding values from the HTTP request.
func ReplaceVariables(str string, req *http.Request) string {
	return varRegexp.ReplaceAllStringFunc(str, func(variable string) string {
		val, err := variableValue(variable, req)
		if err != nil {
			log.Ctx(req.Context()).Debug().Err(err).Msgf("Error replacing variable: %s", variable)
			return variable
		}
		return val
	})
}

// variableValue returns the value of the given NGINX variable based on the HTTP request.
func variableValue(variable string, req *http.Request) (string, error) {
	// $http_name variables are used to access HTTP headers in the request.
	if header, ok := strings.CutPrefix(variable, httpHeaders); ok {
		return strings.Join(req.Header.Values(strings.ReplaceAll(header, "_", "-")), ","), nil
	}

	// $arg_name variables are used to access query parameters in the request URL.
	if arg, ok := strings.CutPrefix(variable, arg); ok {
		return req.URL.Query().Get(arg), nil
	}

	switch variable {
	case host, hostname, bestHTTPHost:
		return req.Host, nil

	case requestURI:
		return req.URL.RequestURI(), nil

	case escapedRequestURI:
		return url.QueryEscape(req.URL.RequestURI()), nil

	case scheme:
		if req.TLS != nil {
			return "https", nil
		}
		return "http", nil

	case args, queryString:
		return req.URL.Query().Encode(), nil

	case remoteAddress:
		return req.RemoteAddr, nil

	default:
		return "", fmt.Errorf("unsupported variable: %s", variable)
	}
}
