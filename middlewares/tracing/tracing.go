package tracing

import (
	"fmt"
	"io"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/tracing/jaeger"
	"github.com/containous/traefik/middlewares/tracing/zipkin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Tracing middleware
type Tracing struct {
	Backend     string         `description:"Selects the tracking backend ('jaeger','zipkin')." export:"true"`
	ServiceName string         `description:"Set the name for this service" export:"true"`
	Jaeger      *jaeger.Config `description:"Settings for jaeger" export:"true"`
	Zipkin      *zipkin.Config `description:"Settings for zipkin" export:"true"`

	opentracing.Tracer
	closer io.Closer
}

// Backend describes things we can use to setup tracing
type Backend interface {
	Setup(serviceName string) (opentracing.Tracer, io.Closer, error)
}

// Setup Tracing middleware
func (t *Tracing) Setup() {
	var err error

	switch t.Backend {
	case jaeger.Name:
		t.Tracer, t.closer, err = t.Jaeger.Setup(t.ServiceName)
	case zipkin.Name:
		t.Tracer, t.closer, err = t.Zipkin.Setup(t.ServiceName)
	default:
		log.Warnf("unknown tracer %q", t.Backend)
		return
	}

	if err != nil {
		log.Warnf("Could not initialize %s tracing: %v", err)
		return
	}

	return
}

// Close tracer
func (t *Tracing) Close() {
	if t.closer != nil {
		t.closer.Close()
	}
}

type statusCodeTracker struct {
	http.ResponseWriter
	status int
}

func (w *statusCodeTracker) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

type epMiddleware struct {
	ep string
	*Tracing
}

func (t *Tracing) NewEntryPoint(name string) *epMiddleware {
	return &epMiddleware{Tracing: t, ep: name}
}

func (t *epMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	opNameFunc := func(r *http.Request) string {
		return fmt.Sprintf("entrypoint %s %s", t.ep, r.Host)
	}

	ctx, _ := t.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := t.StartSpan(opNameFunc(r), ext.RPCServerOption(ctx))
	ext.Component.Set(span, t.ServiceName)
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	span.SetTag("http.host", r.Host)

	w = &statusCodeTracker{w, 200}
	r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))

	next(w, r)

	code := uint16(w.(*statusCodeTracker).status)
	ext.HTTPStatusCode.Set(span, code)
	if code >= 400 {
		ext.Error.Set(span, true)
	}
	span.Finish()
}

type fwdMiddleware struct {
	frontend string
	backend  string
	opName   string
	*Tracing
}

func (t *Tracing) NewForwarder(frontend, backend string) *fwdMiddleware {
	return &fwdMiddleware{
		Tracing:  t,
		frontend: frontend,
		backend:  backend,
		opName:   fmt.Sprintf("forward %s/%s", frontend, backend),
	}
}

func (t *fwdMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := r.Context()
	span, _ := opentracing.StartSpanFromContext(ctx, t.opName)
	span.SetTag("frontend.name", t.frontend)
	span.SetTag("backend.name", t.backend)
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	span.SetTag("http.host", r.Host)

	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	w = &statusCodeTracker{w, 200}
	r = r.WithContext(opentracing.ContextWithSpan(ctx, span))

	next(w, r)

	code := uint16(w.(*statusCodeTracker).status)
	ext.HTTPStatusCode.Set(span, code)
	if code >= 400 {
		ext.Error.Set(span, true)
	}
	span.Finish()
}
