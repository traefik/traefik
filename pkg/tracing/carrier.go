package tracing

import "net/http"

// HTTPHeadersCarrier custom implementation to fix duplicated headers
// It has been fixed in https://github.com/opentracing/opentracing-go/pull/191
type HTTPHeadersCarrier http.Header

// Set conforms to the TextMapWriter interface.
func (c HTTPHeadersCarrier) Set(key, val string) {
	h := http.Header(c)
	h.Set(key, val)
}

// ForeachKey conforms to the TextMapReader interface.
func (c HTTPHeadersCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range c {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
