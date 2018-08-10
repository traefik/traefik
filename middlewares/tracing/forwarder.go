package tracing

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/urfave/negroni"
)

type forwarderMiddleware struct {
	frontend string
	backend  string
	opName   string
	*Tracing
}

// NewForwarderMiddleware creates a new forwarder middleware that traces the outgoing request
func (t *Tracing) NewForwarderMiddleware(frontend, backend string) negroni.Handler {
	log.Debugf("Added outgoing tracing middleware %s", frontend)
	return &forwarderMiddleware{
		Tracing:  t,
		frontend: frontend,
		backend:  backend,
		opName:   generateForwardSpanName(frontend, backend, t.SpanNameLimit),
	}
}

func (f *forwarderMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	span, r, finish := StartSpan(r, f.opName, true)
	defer finish()
	span.SetTag("frontend.name", f.frontend)
	span.SetTag("backend.name", f.backend)
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, fmt.Sprintf("%s%s", r.URL.String(), r.RequestURI))
	span.SetTag("http.host", r.Host)

	InjectRequestHeaders(r)

	recorder := newStatusCodeRecoder(w, 200)

	next(recorder, r)

	LogResponseCode(span, recorder.Status())
}

// generateForwardSpanName will return a Span name of an appropriate lenth based on the 'spanLimit' argument.  If needed, it will be truncated, but will not be less than 21 characters
func generateForwardSpanName(frontend, backend string, spanLimit int) string {
	name := fmt.Sprintf("forward %s/%s", frontend, backend)

	if spanLimit > 0 && len(name) > spanLimit {
		if spanLimit < ForwardMaxLengthNumber {
			log.Warnf("SpanNameLimit is set to be less than required static number of characters, defaulting to %d + 3", ForwardMaxLengthNumber)
			spanLimit = ForwardMaxLengthNumber + 3
		}
		hash := computeHash(name)
		limit := (spanLimit - ForwardMaxLengthNumber) / 2
		name = fmt.Sprintf("forward %s/%s/%s", truncateString(frontend, limit), truncateString(backend, limit), hash)
	}

	return name
}
