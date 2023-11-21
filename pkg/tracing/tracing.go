package tracing

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/tracing/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Backend is an abstraction for tracking backend (OpenTelemetry, ...).
type Backend interface {
	Setup(serviceName string, sampleRate float64, globalAttributes map[string]string, headers map[string]string) (trace.Tracer, io.Closer, error)
}

// Tracer is an abstraction for trace.Tracer.
type Tracer interface {
	trace.Tracer

	GetServiceName() string
	Close()
}

// Tracing middleware.
type Tracing struct {
	trace.Tracer

	spanNameLimit int
	serviceName   string

	closer io.Closer
}

// NewTracing Creates a Tracing.
func NewTracing(conf *static.Tracing) (*Tracing, error) {
	var backend Backend

	if conf.OTLP != nil {
		backend = conf.OTLP
	}

	if backend == nil {
		log.Debug().Msg("Could not initialize tracing, using OpenTelemetry by default")
		defaultBackend := &opentelemetry.Config{}
		backend = defaultBackend
	}

	tracer, closer, err := backend.Setup(conf.ServiceName, conf.SampleRate, conf.GlobalAttributes, conf.Headers)
	if err != nil {
		return nil, err
	}

	return &Tracing{
		spanNameLimit: conf.SpanNameLimit,
		serviceName:   conf.ServiceName,
		Tracer:        tracer,
		closer:        closer,
	}, nil
}

// GetServiceName returns the serviceName.
func (t *Tracing) GetServiceName() string {
	return t.serviceName
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

func Propagator(ctx context.Context, headers http.Header) context.Context {
	propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	return propgator.Extract(ctx, propagation.HeaderCarrier(headers))
}

// LogRequest used to create span tags from the request.
func LogRequest(span trace.Span, r *http.Request, spanKind trace.SpanKind) {
	if span != nil && r != nil {
		// Common attributes https://github.com/open-telemetry/semantic-conventions/blob/v1.23.0/docs/http/http-spans.md#common-attributes
		// TODO: the semconv does not implement Semantic Convention v1.23.0.
		span.SetAttributes(attribute.String("http.request.method", r.Method))
		span.SetAttributes(semconv.NetworkProtocolVersion(proto(r.Proto)))
		// TODO: recommended in OpenTelemetry
		// span.SetAttributes(attribute.String("network.peer.address", "127.0.0.1"))
		// span.SetAttributes(attribute.String("network.peer.port", 80))

		switch spanKind {
		case trace.SpanKindServer:
			logServerRequest(span, r)
		case trace.SpanKindInternal, trace.SpanKindClient:
			logClientRequest(span, r)
		}
	}
}

func logClientRequest(span trace.Span, r *http.Request) {
	// Client attributes https://github.com/open-telemetry/semantic-conventions/blob/v1.23.0/docs/http/http-spans.md#http-client
	span.SetAttributes(attribute.String("http.request.method", r.Method))
	span.SetAttributes(semconv.URLFull(r.URL.String()))
	span.SetAttributes(semconv.URLScheme(r.URL.Scheme))
	span.SetAttributes(semconv.UserAgentOriginal(r.UserAgent()))

	host, sPort, err := net.SplitHostPort(r.URL.Host)
	if err != nil {
		span.SetAttributes(semconv.ServerAddress(r.URL.Host))
		switch r.URL.Scheme {
		case "http":
			span.SetAttributes(semconv.ServerPort(80))
		case "https":
			span.SetAttributes(semconv.ServerPort(443))
		}
	} else {
		port, _ := strconv.Atoi(sPort)
		span.SetAttributes(semconv.ServerAddress(host))
		span.SetAttributes(semconv.ServerPort(port))
	}
}

func logServerRequest(span trace.Span, r *http.Request) {
	// Server attributes https://github.com/open-telemetry/semantic-conventions/blob/v1.23.0/docs/http/http-spans.md#http-server-semantic-conventions
	span.SetAttributes(attribute.Int("http.request.body.size", int(r.ContentLength)))
	span.SetAttributes(semconv.NetworkProtocolVersion(proto(r.Proto)))
	span.SetAttributes(semconv.URLPath(r.URL.Path))
	span.SetAttributes(semconv.URLQuery(r.URL.RawQuery))
	span.SetAttributes(semconv.URLScheme(r.URL.Scheme))
	span.SetAttributes(semconv.UserAgentOriginal(r.UserAgent()))
	span.SetAttributes(semconv.ServerAddress(r.Host))
	// TODO: how to retrieve the entrypoint port?
	// span.SetAttributes(semconv.ServerPort(0))

	rAddr, sPort, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		span.SetAttributes(semconv.ClientAddress(r.RemoteAddr))
	} else {
		span.SetAttributes(semconv.ClientAddress(rAddr))
		port, _ := strconv.Atoi(sPort)
		span.SetAttributes(semconv.ClientPort(port))
	}

	span.SetAttributes(semconv.ClientSocketAddress(r.Header.Get("X-Forwarded-For")))
}

func proto(proto string) string {
	switch proto {
	case "HTTP/1.0":
		return "1.0"
	case "HTTP/1.1":
		return "1.1"
	case "HTTP/2":
		return "2"
	case "HTTP/3":
		return "3"
	default:
		return proto
	}
}

// LogResponseCode used to log response code in span.
func LogResponseCode(span trace.Span, code int) {
	if span != nil {
		span.SetStatus(ServerStatus(code))
		if code > 0 {
			// TODO: the semconv does not implement Semantic Convention v1.23.0.
			span.SetAttributes(attribute.Int("http.response.status_code", code))
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

// InjectRequestHeaders used to inject OpenTelemetry headers into the context.
func InjectRequestHeaders(ctx context.Context, headers http.Header) {
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	propagator.Inject(ctx, propagation.HeaderCarrier(headers))
}

// SetErrorWithEvent flags the span as in error and log an event.
func SetErrorWithEvent(ctx context.Context, format string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetStatus(codes.Error, fmt.Sprintf(format, args...))
	}
}
