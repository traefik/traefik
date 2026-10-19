package failover

import (
	"net/http"

	"github.com/traefik/traefik/v3/pkg/types"
)

type responseWriter struct {
	http.ResponseWriter

	needFallback    bool
	written         bool
	header          http.Header
	statusCodeRange types.HTTPCodeRanges
}

func (r *responseWriter) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}

	if r.needFallback {
		// As we will fallback, we can discard the response body.
		return len(b), nil
	}

	return r.ResponseWriter.Write(b)
}

func (r *responseWriter) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}

	return r.header
}

func (r *responseWriter) WriteHeader(statusCode int) {
	if statusCode >= 100 && statusCode <= 199 && statusCode != http.StatusSwitchingProtocols {
		clear(r.header)

		return
	}

	if !r.written {
		r.written = true
		r.needFallback = r.statusCodeRange.Contains(statusCode)

		if !r.needFallback {
			for k, v := range r.header {
				for _, vv := range v {
					r.ResponseWriter.Header().Add(k, vv)
				}
			}

			r.ResponseWriter.WriteHeader(statusCode)
		}
	}
}

// FlushError sends any buffered data to the client.
func (r *responseWriter) FlushError() error {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}

	// The response is discarded for the fallback,
	// and flushing it would commit an implicit 200 to the client.
	if r.needFallback {
		return nil
	}

	return http.NewResponseController(r.ResponseWriter).Flush()
}

// Unwrap exposes the underlying ResponseWriter for http.ResponseController.
func (r *responseWriter) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
