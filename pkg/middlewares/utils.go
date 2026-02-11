package middlewares

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	scheme            = "$scheme"
	host              = "$host"
	hostname          = "$hostname"
	requestUri        = "$request_uri"
	escapedRequestUri = "$escaped_request_uri"
	path              = "$path"
	args              = "$args"
	remoteAddress     = "$remote_addr"
	// requestMethod     = "$request_method"
	// uri               = "$uri"
	// httpHost          = "$http_host"
)

var nginxVariables = []string{scheme, host, hostname, requestUri, escapedRequestUri, path, args, remoteAddress}

func ReplaceNginxVariables(src string, req *http.Request) string {
	for _, variable := range nginxVariables {
		val := getNginxVariableValue(variable, req)
		if val != "" {
			src = strings.ReplaceAll(src, variable, val)
		}
	}

	return src
}

func getNginxVariableValue(variable string, req *http.Request) string {
	switch variable {
	case host:
		return req.Host
	case requestUri:
		if req.URL != nil {
			return req.URL.RequestURI()
		}
		return req.RequestURI
	case escapedRequestUri:
		if req.URL != nil {
			return regexp.QuoteMeta(req.URL.RequestURI())
		}
		return regexp.QuoteMeta(req.RequestURI)
	case scheme:
		if req.URL == nil {
			return ""
		}
		return req.URL.Scheme
	case path:
		if req.URL == nil {
			return ""
		}
		return req.URL.Path
	case args:
		if req.URL == nil {
			return ""
		}
		return req.URL.RawQuery
	case remoteAddress:
		return req.RemoteAddr
	default:
		return ""
	}
}
