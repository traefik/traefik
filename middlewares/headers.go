package middlewares

//Middleware based on https://github.com/unrolled/secure

import (
	"github.com/containous/traefik/types"
	"net/http"
)

// HeaderOptions is a struct for specifying configuration options for the headers middleware.
type HeaderOptions struct {
	// If Custom request headers are set, these will be added to the request
	CustomRequestHeaders map[string]string
	// If Custom response headers are set, these will be added to the ResponseWriter
	CustomResponseHeaders map[string]string
}

// HeaderStruct is a middleware that helps setup a few basic security features. A single headerOptions struct can be
// provided to configure which features should be enabled, and the ability to override a few of the default values.
type HeaderStruct struct {
	// Customize headers with a headerOptions struct.
	opt HeaderOptions
}

// NewHeaderFromStruct constructs a new header instance from supplied frontend header struct.
func NewHeaderFromStruct(headers types.Headers) *HeaderStruct {
	o := HeaderOptions{
		CustomRequestHeaders:  headers.CustomRequestHeaders,
		CustomResponseHeaders: headers.CustomResponseHeaders,
	}

	return &HeaderStruct{
		opt: o,
	}
}

// NewHeader constructs a new header instance with supplied options.
func NewHeader(options ...HeaderOptions) *HeaderStruct {
	var o HeaderOptions
	if len(options) == 0 {
		o = HeaderOptions{}
	} else {
		o = options[0]
	}

	return &HeaderStruct{
		opt: o,
	}
}

// Handler implements the http.HandlerFunc for integration with the standard net/http lib.
func (s *HeaderStruct) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Let headers process the request.
		s.Process(w, r)
		h.ServeHTTP(w, r)
	})
}

func (s *HeaderStruct) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	s.Process(w, r)
	// If there is a next, call it.
	if next != nil {
		next(w, r)
	}
}

// Process runs the actual checks and returns an error if the middleware chain should stop.
func (s *HeaderStruct) Process(w http.ResponseWriter, r *http.Request) {
	// Loop through Custom request headers
	for header, value := range s.opt.CustomRequestHeaders {
		r.Header.Set(header, value)
	}

	// Loop through Custom response headers
	for header, value := range s.opt.CustomResponseHeaders {
		w.Header().Add(header, value)
	}
}
