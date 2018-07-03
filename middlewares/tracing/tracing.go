package tracing

import (
	"fmt"
	"io"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/tracing/datadog"
	"github.com/containous/traefik/middlewares/tracing/jaeger"
	"github.com/containous/traefik/middlewares/tracing/zipkin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Tracing middleware
type Tracing struct {
	Backend     string          `description:"Selects the tracking backend ('jaeger','zipkin', 'datadog')." export:"true"`
	ServiceName string          `description:"Set the name for this service" export:"true"`
	Jaeger      *jaeger.Config  `description:"Settings for jaeger"`
	Zipkin      *zipkin.Config  `description:"Settings for zipkin"`
	DataDog     *datadog.Config `description:"Settings for DataDog"`

	tracer opentracing.Tracer
	closer io.Closer
}

// StartSpan delegates to opentracing.Tracer
func (t *Tracing) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return t.tracer.StartSpan(operationName, opts...)
}

// Inject delegates to opentracing.Tracer
func (t *Tracing) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return t.tracer.Inject(sm, format, carrier)
}

// Extract delegates to opentracing.Tracer
func (t *Tracing) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return t.tracer.Extract(format, carrier)
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
		t.tracer, t.closer, err = t.Jaeger.Setup(t.ServiceName)
	case zipkin.Name:
		t.tracer, t.closer, err = t.Zipkin.Setup(t.ServiceName)
	case datadog.Name:
		t.tracer, t.closer, err = t.DataDog.Setup(t.ServiceName)
	default:
		log.Warnf("Unknown tracer %q", t.Backend)
		return
	}

	if err != nil {
		log.Warnf("Could not initialize %s tracing: %v", t.Backend, err)
	}
}

// IsEnabled determines if tracing was successfully activated
func (t *Tracing) IsEnabled() bool {
	if t == nil || t.tracer == nil {
		return false
	}
	return true
}

// Close tracer
func (t *Tracing) Close() {
	if t.closer != nil {
		err := t.closer.Close()
		if err != nil {
			log.Warn(err)
		}
	}
}

// LogRequest used to create span tags from the request
func LogRequest(span opentracing.Span, r *http.Request) {
	if span != nil && r != nil {
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())
		span.SetTag("http.host", r.Host)
	}
}

// LogResponseCode used to log response code in span
func LogResponseCode(span opentracing.Span, code int) {
	if span != nil {
		ext.HTTPStatusCode.Set(span, uint16(code))
		if code >= 400 {
			ext.Error.Set(span, true)
		}
	}
}

// GetSpan used to retrieve span from request context
func GetSpan(r *http.Request) opentracing.Span {
	return opentracing.SpanFromContext(r.Context())
}

// InjectRequestHeaders used to inject OpenTracing headers into the request
func InjectRequestHeaders(r *http.Request) {
	if span := GetSpan(r); span != nil {
		err := opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Error(err)
		}
	}
}

// LogEventf logs an event to the span in the request context.
func LogEventf(r *http.Request, format string, args ...interface{}) {
	if span := GetSpan(r); span != nil {
		span.LogKV("event", fmt.Sprintf(format, args...))
	}
}

// StartSpan starts a new span from the one in the request context
func StartSpan(r *http.Request, operationName string, spanKinClient bool, opts ...opentracing.StartSpanOption) (opentracing.Span, *http.Request, func()) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), operationName, opts...)
	if spanKinClient {
		ext.SpanKindRPCClient.Set(span)
	}
	r = r.WithContext(ctx)
	return span, r, func() {
		span.Finish()
	}
}

// SetError flags the span associated with this request as in error
func SetError(r *http.Request) {
	if span := GetSpan(r); span != nil {
		ext.Error.Set(span, true)
	}
}

// SetErrorAndDebugLog flags the span associated with this request as in error and create a debug log
func SetErrorAndDebugLog(r *http.Request, format string, args ...interface{}) {
	SetError(r)
	log.Debugf(format, args...)
	LogEventf(r, format, args...)
}

// SetErrorAndWarnLog flags the span associated with this request as in error and create a debug log
func SetErrorAndWarnLog(r *http.Request, format string, args ...interface{}) {
	SetError(r)
	log.Warnf(format, args...)
	LogEventf(r, format, args...)
}
