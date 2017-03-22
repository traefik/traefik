package audittap

import (
	"bytes"
	audittypes "github.com/containous/traefik/middlewares/audittap/audittypes"
	"net/http"
	"strings"
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

// DropHopByHopHeaders eliminates unintersting headers that only serve to mediate hop-by-hop exchanges.
func (h Headers) DropHopByHopHeaders() Headers {
	delete(h, "connection")
	delete(h, "keep-alive")
	delete(h, "proxy-authenticate")
	delete(h, "proxy-authorization")
	delete(h, "te")
	delete(h, "trailers")
	delete(h, "transfer-encoding")
	delete(h, "upgrade")
	return h
}

func flattenKey(key string) string {
	b := bytes.Buffer{}
	parts := strings.Split(key, "-")
	for i, p := range parts {
		p = strings.ToLower(p)
		if i == 0 || len(p) <= 1 {
			b.WriteString(p)
		} else {
			b.WriteString(strings.ToUpper(p[:1]))
			b.WriteString(p[1:])
		}
	}
	return b.String()
}

// CamelCaseKeys transforms all the keys by removing dashes and using camel-case instead.
func (h Headers) CamelCaseKeys() Headers {
	result := make(http.Header)
	for k, v := range h {
		result[flattenKey(k)] = v
	}
	return Headers(result)
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
func (h Headers) Flatten(prefix string) audittypes.DataMap {
	flat := make(audittypes.DataMap)
	for k, v := range h {
		if len(v) == 1 {
			flat[prefix+k] = v[0]
		} else {
			flat[prefix+k] = v
		}
	}
	return flat
}
