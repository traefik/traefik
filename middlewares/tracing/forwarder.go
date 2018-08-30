package tracing

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
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
		opName:   fmt.Sprintf("forward %s/%s", frontend, backend),
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

	r.Header.Del(jaeger.TracerStateHeaderName)
	InjectRequestHeaders(r)

	recorder := newStatusCodeRecoder(w, 200)

	next(recorder, r)

	LogResponseCode(span, recorder.Status())
}
