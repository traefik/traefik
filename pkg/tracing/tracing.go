package tracing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type contextKey int

const (
	// SpanKindNoneEnum Span kind enum none.
	SpanKindNoneEnum trace.SpanKind = -1
	tracingKey       contextKey     = iota
)

// WithTracing Adds Tracing into the context.
func WithTracing(ctx context.Context, tracing *Tracing) context.Context {
	return context.WithValue(ctx, tracingKey, tracing)
}

// FromContext Gets Tracing from context.
func FromContext(ctx context.Context) (*Tracing, error) {
	if ctx == nil {
		panic("nil context")
	}

	tracer, ok := ctx.Value(tracingKey).(*Tracing)
	if !ok {
		return nil, errors.New("unable to find tracing in the context")
	}
	return tracer, nil
}

// Backend is an abstraction for tracking backend (Jaeger, Zipkin, ...).
type Backend interface {
	Setup(componentName string) (trace.Tracer, io.Closer, error)
}

// Tracing middleware.
type Tracing struct {
	ServiceName   string `description:"Sets the name for this service" export:"true"`
	SpanNameLimit int    `description:"Sets the maximum character limit for span names (default 0 = no limit)" export:"true"`

	tracer trace.Tracer
	closer io.Closer
}

// NewTracing Creates a Tracing.
func NewTracing(serviceName string, spanNameLimit int, tracingBackend Backend) (*Tracing, error) {
	tracing := &Tracing{
		ServiceName:   serviceName,
		SpanNameLimit: spanNameLimit,
	}

	var err error
	tracing.tracer, tracing.closer, err = tracingBackend.Setup(serviceName)
	if err != nil {
		return nil, err
	}
	return tracing, nil
}

// GetTracer returns the tracer.
func (t *Tracing) GetTracer() trace.Tracer {
	return t.tracer
}

// StartSpan delegates to trace.Tracer.
func (t *Tracing) StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, operationName)
}

// StartSpanf delegates to StartSpan.
func (t *Tracing) StartSpanf(r *http.Request, spanKind trace.SpanKind, opPrefix string, opParts []string, separator string, opts ...trace.SpanStartOption) (trace.Span, *http.Request, func()) {
	propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	parentCtx := propgator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	r = r.WithContext(parentCtx)
	operationName := generateOperationName(opPrefix, opParts, separator, t.SpanNameLimit)
	return StartSpan(t.tracer, r, operationName, spanKind, opts...)
}

// IsEnabled determines if tracing was successfully activated.
func (t *Tracing) IsEnabled() bool {
	return t != nil && t.tracer != nil
}

// Close tracer.
func (t *Tracing) Close() {
	if t.closer != nil {
		err := t.closer.Close()
		if err != nil {
			log.Warn().Err(err).Send()
		}
	}
}

// LogRequest used to create span tags from the request.
func LogRequest(span trace.Span, r *http.Request) {
	if span != nil && r != nil {
		span.SetAttributes(semconv.HTTPMethod(r.Method))
		span.SetAttributes(semconv.HTTPURL(r.URL.String()))
		span.SetAttributes(attribute.String("http.host", r.Host))
		span.SetAttributes(semconv.HTTPUserAgent(r.UserAgent()))
		span.SetAttributes(semconv.HTTPRequestContentLength(int(r.ContentLength)))
		clientIP := r.RemoteAddr
		tmp, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			clientIP = tmp
		}
		span.SetAttributes(semconv.HTTPClientIP(clientIP))
	}
}

// LogResponseCode used to log response code in span.
func LogResponseCode(span trace.Span, code int) {
	if span != nil {
		span.SetStatus(ServerStatus(code))
		if code > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(code))
		}
	}
}

// ServerStatus returns a span status code and message for an HTTP status code
// value returned by a server. Status codes in the 400-499 range are not
// returned as errors.
func ServerStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}

// GetSpan used to retrieve span from request context.
func GetSpan(r *http.Request) trace.Span {
	return trace.SpanFromContext(r.Context())
}

// InjectRequestHeaders used to inject OpenTelemetry headers into the request.
func InjectRequestHeaders(r *http.Request) {
	propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	propgator.Inject(r.Context(), propagation.HeaderCarrier(r.Header))
}

// LogEventf logs an event to the span in the request context.
func LogEventf(r *http.Request, format string, args ...interface{}) {
	if span := GetSpan(r); span != nil {
		span.AddEvent(fmt.Sprintf(format, args...))
	}
}

// StartSpan starts a new span from the one in the request context.
func StartSpan(t trace.Tracer, r *http.Request, operationName string, spanKind trace.SpanKind, opts ...trace.SpanStartOption) (trace.Span, *http.Request, func()) {
	ctx, span := t.Start(r.Context(), operationName, opts...)

	switch spanKind {
	case trace.SpanKindProducer, trace.SpanKindConsumer:
		span.SetAttributes(attribute.String("span.kind", spanKind.String()))
	default:
		// noop
	}

	r = r.WithContext(ctx)
	return span, r, func() { span.End() }
}

// SetErrorWithEvent flags the span associated with this request as in error and log an event.
func SetErrorWithEvent(r *http.Request, format string, args ...interface{}) {
	if span := GetSpan(r); span != nil {
		span.SetStatus(codes.Error, fmt.Sprintf(format, args...))
	}
}
