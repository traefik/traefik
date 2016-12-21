package middlewares

import (
	"net/http"
)

// AddPrefix is a middleware used to add prefix to an URL request
type AddPrefix struct {
	Handler http.Handler
	Prefix  string
}

func (s *AddPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = s.Prefix + r.URL.Path
	r.RequestURI = r.URL.RequestURI()
	s.Handler.ServeHTTP(w, r)
}

// SetHandler sets handler
func (s *AddPrefix) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}
