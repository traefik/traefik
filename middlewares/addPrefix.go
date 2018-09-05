package middlewares

import (
	"context"
	"net/http"
)

// AddPrefix is a middleware used to add prefix to an URL request
type AddPrefix struct {
	Handler http.Handler
	Prefix  string
}

type key string

const (
	// AddPrefixKey is the key within the request context used to
	// store the added prefix
	AddPrefixKey key = "AddPrefix"
)

func (s *AddPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = s.Prefix + r.URL.Path
	if r.URL.RawPath != "" {
		r.URL.RawPath = s.Prefix + r.URL.RawPath
	}
	r.RequestURI = r.URL.RequestURI()
	r = r.WithContext(context.WithValue(r.Context(), AddPrefixKey, s.Prefix))
	s.Handler.ServeHTTP(w, r)
}

// SetHandler sets handler
func (s *AddPrefix) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}
