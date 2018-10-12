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
		if shouldLogPanic(err) {
			log.Errorf("Recovered from panic in http handler: %+v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			log.Debugf("HTTP handler has been aborted: %v", err)
		}
	}
}

// https://github.com/golang/go/blob/c33153f7b416c03983324b3e8f869ce1116d84bc/src/net/http/httputil/reverseproxy.go#L284
func shouldLogPanic(panicValue interface{}) bool {
	return panicValue != nil && panicValue != http.ErrAbortHandler
}
