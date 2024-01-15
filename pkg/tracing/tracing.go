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

// NewTracing Creates a Tracing.
func NewTracing(conf *static.Tracing) (trace.Tracer, io.Closer, error) {
	var backend Backend

	if conf.OTLP != nil {
		backend = conf.OTLP
	}

	if backend == nil {
		log.Debug().Msg("Could not initialize tracing, using OpenTelemetry by default")
		defaultBackend := &opentelemetry.Config{}
		backend = defaultBackend
	}

	return backend.Setup(conf.ServiceName, conf.SampleRate, conf.GlobalAttributes, conf.Headers)
}

// TracerFromContext extracts the trace.Tracer from the given context.
func TracerFromContext(ctx context.Context) trace.Tracer {
	span := trace.SpanFromContext(ctx)
	if span != nil && span.TracerProvider() != nil {
		return span.TracerProvider().Tracer("github.com/traefik/traefik")
	}

	return nil
}

// ExtractCarrierIntoContext reads cross-cutting concerns from the carrier into a Context.
func ExtractCarrierIntoContext(ctx context.Context, headers http.Header) context.Context {
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	return propagator.Extract(ctx, propagation.HeaderCarrier(headers))
}

// InjectContextIntoCarrier sets cross-cutting concerns from the request context into the request headers.
func InjectContextIntoCarrier(req *http.Request) {
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	propagator.Inject(req.Context(), propagation.HeaderCarrier(req.Header))
}

// SetStatusErrorf flags the span as in error and log an event.
func SetStatusErrorf(ctx context.Context, format string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetStatus(codes.Error, fmt.Sprintf(format, args...))
	}
}

// LogClientRequest used to add span attributes from the request as a Client.
// TODO: the semconv does not implement Semantic Convention v1.23.0.
func LogClientRequest(span trace.Span, r *http.Request) {
	if r == nil || span == nil {
		return
	}

	// Common attributes https://github.com/open-telemetry/semantic-conventions/blob/v1.23.0/docs/http/http-spans.md#common-attributes
	span.SetAttributes(semconv.HTTPRequestMethodKey.String(r.Method))
	span.SetAttributes(semconv.NetworkProtocolVersion(proto(r.Proto)))

	// Client attributes https://github.com/open-telemetry/semantic-conventions/blob/v1.23.0/docs/http/http-spans.md#http-client
	span.SetAttributes(semconv.URLFull(r.URL.String()))
	span.SetAttributes(semconv.URLScheme(r.URL.Scheme))
	span.SetAttributes(semconv.UserAgentOriginal(r.UserAgent()))

	host, port, err := net.SplitHostPort(r.URL.Host)
	if err != nil {
		span.SetAttributes(attribute.String("network.peer.address", host))
		span.SetAttributes(semconv.ServerAddress(r.URL.Host))
		switch r.URL.Scheme {
		case "http":
			span.SetAttributes(attribute.String("network.peer.port", "80"))
			span.SetAttributes(semconv.ServerPort(80))
		case "https":
			span.SetAttributes(attribute.String("network.peer.port", "443"))
			span.SetAttributes(semconv.ServerPort(443))
		}
	} else {
		span.SetAttributes(attribute.String("network.peer.address", host))
		span.SetAttributes(attribute.String("network.peer.port", port))
		intPort, _ := strconv.Atoi(port)
		span.SetAttributes(semconv.ServerAddress(host))
		span.SetAttributes(semconv.ServerPort(intPort))
	}
}

// LogServerRequest used to add span attributes from the request as a Server.
// TODO: the semconv does not implement Semantic Convention v1.23.0.
func LogServerRequest(span trace.Span, r *http.Request) {
	if r == nil {
		return
	}

	// Common attributes https://github.com/open-telemetry/semantic-conventions/blob/v1.23.0/docs/http/http-spans.md#common-attributes
	span.SetAttributes(semconv.HTTPRequestMethodKey.String(r.Method))
	span.SetAttributes(semconv.NetworkProtocolVersion(proto(r.Proto)))

	// Server attributes https://github.com/open-telemetry/semantic-conventions/blob/v1.23.0/docs/http/http-spans.md#http-server-semantic-conventions
	span.SetAttributes(semconv.HTTPRequestBodySize(int(r.ContentLength)))
	span.SetAttributes(semconv.URLPath(r.URL.Path))
	span.SetAttributes(semconv.URLQuery(r.URL.RawQuery))
	span.SetAttributes(semconv.URLScheme(r.Header.Get("X-Forwarded-Proto")))
	span.SetAttributes(semconv.UserAgentOriginal(r.UserAgent()))
	span.SetAttributes(semconv.ServerAddress(r.Host))

	host, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		span.SetAttributes(semconv.ClientAddress(r.RemoteAddr))
		span.SetAttributes(attribute.String("network.peer.address", r.RemoteAddr))
	} else {
		span.SetAttributes(attribute.String("network.peer.address", host))
		span.SetAttributes(attribute.String("network.peer.port", port))
		span.SetAttributes(semconv.ClientAddress(host))
		intPort, _ := strconv.Atoi(port)
		span.SetAttributes(semconv.ClientPort(intPort))
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
func LogResponseCode(span trace.Span, code int, spanKind trace.SpanKind) {
	if span != nil {
		var status codes.Code
		var desc string
		switch spanKind {
		case trace.SpanKindServer:
			status, desc = ServerStatus(code)
		case trace.SpanKindClient:
			status, desc = ClientStatus(code)
		default:
			status, desc = DefaultStatus(code)
		}
		span.SetStatus(status, desc)
		if code > 0 {
			span.SetAttributes(semconv.HTTPResponseStatusCode(code))
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

// ClientStatus returns a span status code and message for an HTTP status code
// value returned by a server. Status codes in the 400-499 range are not
// returned as errors.
func ClientStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 400 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}

// DefaultStatus returns a span status code and message for an HTTP status code
// value generated internally.
func DefaultStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}
