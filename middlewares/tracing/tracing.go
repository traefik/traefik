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
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/urfave/negroni"
)

// Tracing middleware
type Tracing struct {
	Backend     string         `description:"Selects the tracking backend ('jaeger','zipkin')." export:"true"`
	ServiceName string         `description:"Set the name for this service" export:"true"`
	Jaeger      *jaeger.Config `description:"Settings for jaeger"`
	Zipkin      *zipkin.Config `description:"Settings for zipkin"`

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

// IsEnabled determines if tracing was successfully activated
func (t *Tracing) IsEnabled() bool {
	if t == nil || t.Tracer == nil {
		return false
	}
	return true
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

func (s *statusCodeTracker) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

type entryPointMiddleware struct {
	entryPoint string
	*Tracing
}

// NewEntryPoint creates a new middleware that the incoming request
func (t *Tracing) NewEntryPoint(name string) negroni.Handler {
	return &entryPointMiddleware{Tracing: t, entryPoint: name}
}

func (e *entryPointMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	opNameFunc := func(r *http.Request) string {
		return fmt.Sprintf("entrypoint %s %s", e.entryPoint, r.Host)
	}

	ctx, _ := e.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := e.StartSpan(opNameFunc(r), ext.RPCServerOption(ctx))
	ext.Component.Set(span, e.ServiceName)
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	span.SetTag("http.host", r.Host)
	ext.SpanKindRPCServer.Set(span)

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

// NewForwarder creates a new forwarder middleware that the final outgoing request
func (t *Tracing) NewForwarder(frontend, backend string) negroni.Handler {
	return &fwdMiddleware{
		Tracing:  t,
		frontend: frontend,
		backend:  backend,
		opName:   fmt.Sprintf("forward %s/%s", frontend, backend),
	}
}

func (f *fwdMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := r.Context()
	span, _ := opentracing.StartSpanFromContext(ctx, f.opName)
	span.SetTag("frontend.name", f.frontend)
	span.SetTag("backend.name", f.backend)
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	span.SetTag("http.host", r.Host)
	ext.SpanKindRPCClient.Set(span)

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

// LogEventf logs an event to the span in the request context.
func LogEventf(r *http.Request, str string, vals ...interface{}) {
	if span := opentracing.SpanFromContext(r.Context()); span != nil {
		span.LogKV("event", fmt.Sprintf(str, vals...))
	}
}

// DebugEventf logs an event to the span in the request context, and additionally logs
// to debug logging.
func DebugEventf(r *http.Request, str string, vals ...interface{}) {
	log.Debugf(str, vals...)
	if span := opentracing.SpanFromContext(r.Context()); span != nil {
		span.LogKV("event", fmt.Sprintf(str, vals...))
	}
}

// LogFields logs the opentracing log fields to the span in the request context.
func LogFields(r *http.Request, fields ...otlog.Field) {
	if span := opentracing.SpanFromContext(r.Context()); span != nil {
		span.LogFields(fields...)
	}
}

// StartSpan starts a new span from the one in the request context
func StartSpan(r *http.Request, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), operationName, opts...)
	r = r.WithContext(ctx)
	return span, r
}

type baseMiddleware struct {
	opName string
	*Tracing
}

func (t *baseMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	span, nr := StartSpan(r, t.opName)
	next(w, nr)
	span.Finish()
}

// SetError flags the span associated with this request as in error
func SetError(r *http.Request) {
	if span := opentracing.SpanFromContext(r.Context()); span != nil {
		ext.Error.Set(span, true)
	}
}
