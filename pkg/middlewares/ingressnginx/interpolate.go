package ingressnginx

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	scheme            = "$scheme"
	host              = "$host"
	httpHeaders       = "$http_"
	bestHTTPHost      = "$best_http_host"
	hostname          = "$hostname"
	requestURI        = "$request_uri"
	escapedRequestURI = "$escaped_request_uri"
	path              = "$path"
	args              = "$args"
	arg               = "$arg_"
	remoteAddress     = "$remote_addr"
)

func ReplaceNginxVariables(src string, req *http.Request) string {
	varsRegexp := regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`)
	results := varsRegexp.FindAllString(src, -1)

	for _, variable := range results {
		val := getNginxVariableValue(variable, req)
		if val != "" {
			src = strings.ReplaceAll(src, variable, val)
		}
	}

	return src
}

func getNginxVariableValue(variable string, req *http.Request) string {
	if header, ok := strings.CutPrefix(variable, httpHeaders); ok {
		return strings.Join(req.Header.Values(strings.ReplaceAll(header, "_", "-")), ",")
	}

	if arg, ok := strings.CutPrefix(variable, arg); ok {
		return req.URL.Query().Get(arg)
	}

	switch variable {
	case host, hostname, bestHTTPHost:
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
		return req.URL.Query().Encode()
	case remoteAddress:
		return req.RemoteAddr
	default:
		return ""
	}
}

// for NGINX compatibility on auth-signin
func UpdateAuthSigninURL(src string) string {
	if strings.Contains(src, "rd=") {
		return src
	}
	suffix := "rd=$scheme://$host$escaped_request_uri"
	if !strings.Contains(src, "?") {
		suffix = "?" + suffix
	} else {
		suffix = "&" + suffix
	}
	return src + suffix
}
