package ingressnginx

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// This the list of supported NGINX variables for interpolation.
// It is not exhaustive, but covers the most commonly used ones in Ingress NGINX annotations.
const (
	scheme        = "scheme"
	host          = "host"
	httpHeaders   = "http_"
	hostname      = "hostname"
	requestURI    = "request_uri"
	requestMethod = "request_method"
	queryString   = "query_string"
	args          = "args"
	arg           = "arg_"
	remoteAddress = "remote_addr"
	uri           = "uri"
	documentURI   = "document_uri"
	serverName    = "server_name"
	serverPort    = "server_port"
	contentType   = "content_type"
	contentLength = "content_length"
	cookie        = "cookie_"
	isArgs        = "is_args"

	// Variables set by ingress-nginx template.
	bestHTTPHost          = "best_http_host"
	escapedRequestURI     = "escaped_request_uri"
	proxyAddXForwardedFor = "proxy_add_x_forwarded_for"
)

// varRegexp is a regular expression to match NGINX variables in the form of $variable or $variable_name,
// or capture group references $1-$9.
var varRegexp = regexp.MustCompile(`(\$\{?([a-zA-Z_][a-zA-Z0-9_]*|[1-9])}?)`)

// ReplaceVariables replaces NGINX variables in the given string with their corresponding values from the HTTP request.
// Today this supports the `$scheme`, `$host`, `$http_*`, `$best_http_host`, `$hostname`, `$request_uri`,
// `$escaped_request_uri`, `$query_string`, `$args`, `$arg_*`, `$remote_addr`, `$request_method`,
// `$uri`, `$document_uri`, `$server_name`, `$server_port`, `$content_type`, `$content_length`,
// `$cookie_*`, `$is_args`, and `$proxy_add_x_forwarded_for` variables.
// Custom variables can be passed through the vars param.
func ReplaceVariables(str string, req *http.Request, vars map[string]string) string {
	return varRegexp.ReplaceAllStringFunc(str, func(variable string) string {
		groups := varRegexp.FindStringSubmatch(variable)
		val, err := variableValue(groups[1], groups[2], req, vars)
		if err != nil {
			log.Ctx(req.Context()).Debug().Err(err).Msgf("Error replacing variable: %s", variable)
			return variable
		}
		return val
	})
}

// variableValue returns the value of the given NGINX variable based on the HTTP request and the custom vars map.
func variableValue(rawVariable, variable string, req *http.Request, vars map[string]string) (string, error) {
	// $http_name variables are used to access HTTP headers in the request.
	if header, ok := strings.CutPrefix(variable, httpHeaders); ok {
		return strings.Join(req.Header.Values(strings.ReplaceAll(header, "_", "-")), ","), nil
	}

	// $arg_name variables are used to access query parameters in the request URL.
	if arg, ok := strings.CutPrefix(variable, arg); ok {
		return req.URL.Query().Get(arg), nil
	}

	// $cookie_name variables are used to access cookie values in the request.
	if name, ok := strings.CutPrefix(variable, cookie); ok {
		c, _ := req.Cookie(name)
		if c == nil {
			return "", nil
		}
		return c.Value, nil
	}

	switch variable {
	case host:
		// NGINX's $host returns the hostname without port, lowercased.
		if hostOnly, _, err := net.SplitHostPort(req.Host); err == nil {
			return strings.ToLower(hostOnly), nil
		}
		return strings.ToLower(req.Host), nil

	case bestHTTPHost:
		// ingress-nginx's $best_http_host preserves the port.
		return req.Host, nil

	case hostname:
		h, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("getting hostname: %w", err)
		}
		return h, nil

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
		return req.URL.RawQuery, nil

	case remoteAddress:
		return stripPort(req.RemoteAddr), nil

	case proxyAddXForwardedFor:
		clientIP := stripPort(req.RemoteAddr)
		if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
			return prior + ", " + clientIP, nil
		}
		return clientIP, nil

	case requestMethod:
		return req.Method, nil

	case uri, documentURI:
		return req.URL.Path, nil

	case serverName:
		if hostOnly, _, err := net.SplitHostPort(req.Host); err == nil {
			return hostOnly, nil
		}
		return req.Host, nil

	case serverPort:
		if _, port, err := net.SplitHostPort(req.Host); err == nil {
			return port, nil
		}
		if req.TLS != nil {
			return "443", nil
		}
		return "80", nil

	case contentType:
		return req.Header.Get("Content-Type"), nil

	case contentLength:
		return req.Header.Get("Content-Length"), nil

	case isArgs:
		if req.URL.RawQuery != "" {
			return "?", nil
		}
		return "", nil

	default:
		// Here we are adding the $ prefix back to the variable name to look it up in the custom vars map, as it has been removed by the regular expression.
		// Note that custom variable keys are stored with the $ prefix in the vars map, e.g. {"$my_var": "value"}.
		if value, ok := vars["$"+variable]; ok {
			return value, nil
		}

		return "", fmt.Errorf("unsupported variable: %s", rawVariable)
	}
}

// stripPort removes the port from a host:port address.
// It handles both IPv4 (192.168.1.1:8080) and IPv6 ([::1]:8080) formats.
func stripPort(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
