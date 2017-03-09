package linode

import (
	"bytes"
	"net/url"
	"sort"
)

type (
	// Parameters represents the parameters that are passed for an API call.
	Parameters map[string]string
)

// Get gets the parameter associated with the given key.  If no parameter is
// associated with the key, Get returns the empty string.
func (p Parameters) Get(key string) string {
	if p == nil {
		return ""
	}
	value, ok := p[key]
	if !ok {
		return ""
	}
	return value
}

// Set sets the key to value.  It replaces any existing value.
func (p Parameters) Set(key, value string) {
	p[key] = value
}

// Del deletes the parameter associated with key.
func (p Parameters) Del(key string) {
	delete(p, key)
}

// Encode encodes the parameters into "URL encoded" form ("bar=baz&foo=quux")
// sorted by key.
func (p Parameters) Encode() string {
	var buf bytes.Buffer
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k) + "=")
		buf.WriteString(url.QueryEscape(p[k]))
	}
	return buf.String()
}
