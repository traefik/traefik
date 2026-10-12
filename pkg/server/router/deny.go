package router

import (
	"net/http"

	"github.com/traefik/traefik/v2/pkg/log"
)

// denyConnect rejects the CONNECT requests, whatever the HTTP version.
// CONNECT tunnels are not supported, and forwarding them without a dedicated handling is unsafe.
func denyConnect(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodConnect {
			log.FromContext(req.Context()).Debug("Rejecting CONNECT request: CONNECT tunnels are not supported")
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}

		h.ServeHTTP(rw, req)
	})
}

// denyEncodedPathCharacters reject the request if the escaped path contains encoded characters in the given list.
func denyEncodedPathCharacters(encodedCharacters map[string]struct{}, h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if len(encodedCharacters) == 0 {
			h.ServeHTTP(rw, req)
			return
		}

		escapedPath := req.URL.EscapedPath()

		for i := 0; i < len(escapedPath); i++ {
			if escapedPath[i] != '%' {
				continue
			}

			// This should never happen as the standard library will reject requests containing invalid percent-encodings.
			// This discards URLs with a percent character at the end.
			if i+2 >= len(escapedPath) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			// This rejects a request with a path containing the given encoded characters.
			if _, exists := encodedCharacters[escapedPath[i:i+3]]; exists {
				log.WithoutContext().Debugf("Rejecting request because it contains encoded character %s in the URL path: %s", escapedPath[i:i+3], escapedPath)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			i += 2
		}

		h.ServeHTTP(rw, req)
	})
}
