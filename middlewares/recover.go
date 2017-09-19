package middlewares

import (
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/urfave/negroni"
)

// RecoverHandler recovers from a panic in http handlers
func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer recoverFunc(w)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// NegroniRecoverHandler recovers from a panic in negroni handlers
func NegroniRecoverHandler() negroni.Handler {
	fn := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		defer recoverFunc(w)
		next.ServeHTTP(w, r)
	}
	return negroni.HandlerFunc(fn)
}

func recoverFunc(w http.ResponseWriter) {
	if err := recover(); err != nil {
		log.Errorf("Recovered from panic in http handler: %+v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
