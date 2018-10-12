package middlewares

import (
	"net/http"
	"runtime"

	"github.com/containous/traefik/log"
	"github.com/urfave/negroni"
)

// RecoverHandler recovers from a panic in http handlers
func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer recoverFunc(w, r)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// NegroniRecoverHandler recovers from a panic in negroni handlers
func NegroniRecoverHandler() negroni.Handler {
	fn := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		defer recoverFunc(w, r)
		next.ServeHTTP(w, r)
	}
	return negroni.HandlerFunc(fn)
}

func recoverFunc(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		if !shouldLogPanic(err) {
			log.Debugf("HTTP handler has been aborted [%s %s]: %v", r.RemoteAddr, r.URL, err)
			return
		}

		log.Errorf("Recovered from panic in http handler [%s %s]: %+v", r.RemoteAddr, r.URL, err)

		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		log.Errorf("Stack: %s", buf)

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// https://github.com/golang/go/blob/c33153f7b416c03983324b3e8f869ce1116d84bc/src/net/http/httputil/reverseproxy.go#L284
func shouldLogPanic(panicValue interface{}) bool {
	return panicValue != nil && panicValue != http.ErrAbortHandler
}
