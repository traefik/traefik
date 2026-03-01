package router

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

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
