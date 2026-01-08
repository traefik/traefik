package canonicalpath

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// Config configures the canonical path middleware behavior.
type Config struct {
	// Strategy determines how paths are canonicalized.
	// Default: StrategyPreserveReserved (maintains RFC 3986 semantics)
	Strategy NormalizationStrategy

	// RejectNullByte returns 400 for paths containing null bytes.
	// Null bytes in paths are never legitimate and indicate an attack.
	// Default: true
	RejectNullByte bool

	// RejectMalformedEncoding returns 400 for malformed percent-encoding.
	// Default: true
	RejectMalformedEncoding bool

	// LogRejections logs rejected requests at debug level.
	// Default: true
	LogRejections bool
}

// DefaultConfig returns a secure default configuration.
// Uses StrategyPreserveReserved to maintain RFC 3986 semantics.
func DefaultConfig() Config {
	return Config{
		Strategy:                StrategyPreserveReserved,
		RejectNullByte:          true,
		RejectMalformedEncoding: true,
		LogRejections:           true,
	}
}

// Middleware creates the canonical path middleware.
// This middleware MUST be first in the handler chain to establish
// the canonical path representation used by all subsequent handlers.
//
// The middleware:
// 1. Optionally rejects requests with null bytes or malformed encoding
// 2. Computes the canonical path representation
// 3. Stores both original and canonical paths in the request context
// 4. All downstream handlers use GetCanonical() for routing decisions
// 5. The original path is preserved for backend forwarding
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rawPath := req.URL.EscapedPath()
			logger := log.Ctx(req.Context())

			// Security check: reject malformed percent-encoding
			if cfg.RejectMalformedEncoding && HasMalformedEncoding(rawPath) {
				if cfg.LogRejections {
					logger.Debug().
						Str("path", rawPath).
						Msg("Rejecting request with malformed percent-encoding")
				}
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			// Compute canonical path representation
			pr := Canonicalize(rawPath, cfg.Strategy)

			// Security check: reject null bytes
			if cfg.RejectNullByte && pr.ContainsNullByte() {
				if cfg.LogRejections {
					logger.Debug().
						Str("path", rawPath).
						Msg("Rejecting request with null byte in path")
				}
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			// Store in context for all downstream handlers
			ctx := WithPathRepresentation(req.Context(), pr)

			// Debug logging for troubleshooting
			logger.Trace().
				Str("original_path", pr.Original).
				Str("canonical_path", pr.Canonical).
				Msg("Canonical path established")

			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

// SimpleMiddleware creates the middleware with default configuration.
func SimpleMiddleware(next http.Handler) http.Handler {
	return Middleware(DefaultConfig())(next)
}
