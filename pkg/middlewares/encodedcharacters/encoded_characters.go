package encodedcharacters

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "EncodedCharacters"

type encodedCharacters struct {
	next             http.Handler
	deniedCharacters map[string]struct{}
	name             string
}

// NewEncodedCharacters creates an Encoded Characters middleware.
func NewEncodedCharacters(ctx context.Context, next http.Handler, config dynamic.EncodedCharacters, name string) http.Handler {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	return &encodedCharacters{
		next:             next,
		deniedCharacters: mapDeniedCharacters(config),
		name:             name,
	}
}

func (ec *encodedCharacters) GetTracingInformation() (string, string) {
	return ec.name, typeName
}

func (ec *encodedCharacters) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), ec.name, typeName)

	if len(ec.deniedCharacters) == 0 {
		ec.next.ServeHTTP(rw, req)
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
		if _, exists := ec.deniedCharacters[escapedPath[i:i+3]]; exists {
			logger.Debug().Msgf("Rejecting request because it contains encoded character %s in the URL path: %s", escapedPath[i:i+3], escapedPath)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		i += 2
	}

	ec.next.ServeHTTP(rw, req)
}

// mapDeniedCharacters returns a map of unallowed encoded characters.
func mapDeniedCharacters(config dynamic.EncodedCharacters) map[string]struct{} {
	characters := make(map[string]struct{})

	if !config.AllowEncodedSlash {
		characters["%2F"] = struct{}{}
		characters["%2f"] = struct{}{}
	}
	if !config.AllowEncodedBackSlash {
		characters["%5C"] = struct{}{}
		characters["%5c"] = struct{}{}
	}
	if !config.AllowEncodedNullCharacter {
		characters["%00"] = struct{}{}
	}
	if !config.AllowEncodedSemicolon {
		characters["%3B"] = struct{}{}
		characters["%3b"] = struct{}{}
	}
	if !config.AllowEncodedPercent {
		characters["%25"] = struct{}{}
	}
	if !config.AllowEncodedQuestionMark {
		characters["%3F"] = struct{}{}
		characters["%3f"] = struct{}{}
	}
	if !config.AllowEncodedHash {
		characters["%23"] = struct{}{}
	}

	return characters
}
