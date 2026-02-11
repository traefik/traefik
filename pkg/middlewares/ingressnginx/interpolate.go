package ingressnginx

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	scheme            = "$scheme"
	host              = "$host"
	bestHttpHost      = "$best_http_host"
	hostname          = "$hostname"
	requestURI        = "$request_uri"
	escapedRequestURI = "$escaped_request_uri"
	path              = "$path"
	args              = "$args"
	remoteAddress     = "$remote_addr"
)

var nginxVariables = []string{scheme, host, hostname, requestURI, escapedRequestURI, path, args, remoteAddress}

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
	case host, hostname, bestHttpHost:
		return req.Host
	case requestURI:
		if req.URL != nil {
			return req.URL.RequestURI()
		}
		return req.RequestURI
	case escapedRequestURI:
		if req.URL != nil {
			return url.QueryEscape(req.URL.RequestURI())
		}
		return url.QueryEscape(req.RequestURI)
	case scheme:
		// Determine scheme: check TLS first, then X-Forwarded-Proto header
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		} else if proto := req.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		return scheme
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
