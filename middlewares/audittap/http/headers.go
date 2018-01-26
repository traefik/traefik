package http

import (
	"net/http"
	"strings"

	"github.com/containous/traefik/middlewares/audittap/types"
)

// Headers is a map equivalent to http.Header with some transformation methods.
type Headers http.Header

// NewHeaders creates a new Headers as a copy in which all keys are lowercase.
func NewHeaders(h http.Header) Headers {
	result := make(Headers)
	for k, v := range h {
		result[strings.ToLower(k)] = v
	}
	return result
}

func expandCookies(existing interface{}, v []string) []string {
	var all []string
	if existing != nil {
		all = existing.([]string)
	}
	for _, s := range v {
		cookies := strings.Split(s, ";")
		for _, c := range cookies {
			all = append(all, strings.TrimSpace(c))
		}
	}
	return all
}

// SimplifyCookies splits multi-valued cookie entries and then joins all cookies as a single entry.
func (h Headers) SimplifyCookies() Headers {
	result := make(http.Header)
	for k, v := range h {
		if k == "cookie" {
			result[k] = expandCookies(result[k], v)
		} else {
			result[k] = v
		}
	}
	return Headers(result)
}

// Flatten replaces length=1 []string values with a single string.
// Values with longer length remain unchanged.
func (h Headers) Flatten() types.DataMap {
	flat := make(types.DataMap)
	for k, v := range h {
		if len(v) == 1 {
			flat[k] = v[0]
		} else {
			flat[k] = v
		}
	}
	return flat
}

// ClientAndRequestHeaders populates separate data maps for the client and request
// headers that we want to audit.
func (h Headers) ClientAndRequestHeaders() (clientHeaders, requestHeaders types.DataMap) {
	clientHeaders = make(types.DataMap)
	requestHeaders = make(types.DataMap)

	for k, v := range h {
		var fv interface{}
		if len(v) == 1 {
			fv = v[0]
		} else {
			fv = v
		}

		if headerInSet(k, []string{"x-request-id", "content-type", "true-client-ip", "true-client-port", "x-source", "authorization"}) {
			// Skip as we've recorded this individually
		} else if headerInSet(k, []string{"x-", "forwarded-", "if-", "proxy-", "akamai-"}) {
			requestHeaders[k] = fv
		} else {
			clientHeaders[k] = fv
		}
	}

	return clientHeaders, requestHeaders
}

// ResponseHeaders returns the response headers we are interested in
// as a DataMap
func (h Headers) ResponseHeaders() types.DataMap {
	responseHeaders := make(types.DataMap)

	for k, v := range h {
		var fv interface{}
		if len(v) == 1 {
			fv = v[0]
		} else {
			fv = v
		}

		if headerInSet(k, []string{"x-request-id", "content-type"}) {
			// Skip as we've recorded this individually
		} else {
			(responseHeaders)[k] = fv
		}
	}

	return responseHeaders
}

func headerInSet(hdr string, set []string) bool {
	for _, v := range set {
		if strings.HasPrefix(hdr, v) {
			return true
		}
	}
	return false
}
