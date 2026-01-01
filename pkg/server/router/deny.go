package router

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// denyFragment rejects the request if the URL path contains a fragment (hash character).
// When go receives an HTTP request, it assumes the absence of fragment URL.
// However, it is still possible to send a fragment in the request.
// In this case, Traefik will encode the '#' character, altering the request's intended meaning.
// To avoid this behavior, the following function rejects requests that include a fragment in the URL.
func denyFragment(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.RawPath, "#") {
			log.Debug().Msgf("Rejecting request because it contains a fragment in the URL path: %s", req.URL.RawPath)
			rw.WriteHeader(http.StatusBadRequest)

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
				log.Debug().Msgf("Rejecting request because it contains encoded character %s in the URL path: %s", escapedPath[i:i+3], escapedPath)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			i += 2
		}

		h.ServeHTTP(rw, req)
	})
}
