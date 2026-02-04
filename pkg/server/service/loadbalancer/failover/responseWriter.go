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
	statusCoderange types.HTTPCodeRanges
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
	if statusCode < 200 {
		// forwarding informational status code.
		for k, v := range r.header {
			for _, vv := range v {
				r.ResponseWriter.Header().Add(k, vv)
			}
		}

		r.ResponseWriter.WriteHeader(statusCode)

		return
	}

	if !r.written {
		r.written = true
		r.needFallback = r.statusCoderange.Contains(statusCode)

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

func (r *responseWriter) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
