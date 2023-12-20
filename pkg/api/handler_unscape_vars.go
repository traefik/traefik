package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func UnscapeVarsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requestVars := mux.Vars(r)

		for name, value := range requestVars {
			unscapedValue, err := url.PathUnescape(value)
			if err != nil {
				writeError(rw, fmt.Sprintf("unable to decode %s: %s", name, value), http.StatusBadRequest)
				return
			}

			log.Trace().Msg(fmt.Sprintf("var %s decoded: %s [%s]", name, unscapedValue, value))
			requestVars[name] = unscapedValue
		}

		h.ServeHTTP(rw, r)
	})
}
