package middlewares

import (
	"net/http"

	"github.com/containous/traefik/middlewares/common"
)

// AddPrefix is a middleware used to add prefix to an URL request
type AddPrefix struct {
	common.BasicMiddleware
	Prefix string
}

var _ common.Middleware = &AddPrefix{}

func (s *AddPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = s.Prefix + r.URL.Path
	r.RequestURI = r.URL.RequestURI()
	s.Next().ServeHTTP(w, r)
}
