package observability

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/semconv/v1.37.0/httpconv"
)

const (
	semConvServerMetricsTypeName = "SemConvServerMetrics"
)

type semConvServerMetrics struct {
	next                  http.Handler
	semConvMetricRegistry *metrics.SemConvMetricsRegistry
}

// SemConvServerMetricsHandler return the alice.Constructor for semantic conventions servers metrics.
func SemConvServerMetricsHandler(ctx context.Context, semConvMetricRegistry *metrics.SemConvMetricsRegistry) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newServerMetricsSemConv(ctx, semConvMetricRegistry, next), nil
	}
}

// newServerMetricsSemConv creates a new semConv server metrics middleware for incoming requests.
func newServerMetricsSemConv(ctx context.Context, semConvMetricRegistry *metrics.SemConvMetricsRegistry, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", semConvServerMetricsTypeName).Debug().Msg("Creating middleware")

	return &semConvServerMetrics{
		semConvMetricRegistry: semConvMetricRegistry,
		next:                  next,
	}
}

func (e *semConvServerMetrics) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if e.semConvMetricRegistry == nil || !SemConvMetricsEnabled(req.Context()) {
		e.next.ServeHTTP(rw, req)
		return
	}

	start := time.Now()
	e.next.ServeHTTP(rw, req)
	end := time.Now()

	ctx := req.Context()
	capt, err := capture.FromContext(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str(logs.MiddlewareType, semConvServerMetricsTypeName).Msg("Could not get Capture")
		return
	}

	var attrs []attribute.KeyValue

	if capt.StatusCode() < 100 || capt.StatusCode() >= 600 {
		attrs = append(attrs, attribute.Key("error.type").String(fmt.Sprintf("Invalid HTTP status code ; %d", capt.StatusCode())))
	} else if capt.StatusCode() >= 400 {
		attrs = append(attrs, attribute.Key("error.type").String(strconv.Itoa(capt.StatusCode())))
	}

	// Additional optional attributes.
	attrs = append(attrs, semconv.HTTPResponseStatusCode(capt.StatusCode()))
	attrs = append(attrs, semconv.NetworkProtocolName(strings.ToLower(req.Proto)))
	attrs = append(attrs, semconv.NetworkProtocolVersion(Proto(req.Proto)))

	if route, ok := HTTPRouteFromContext(req.Context()); ok {
		attrs = append(attrs, semconv.HTTPRoute(route))
	}

	if addr := serverAddress(req); addr != "" {
		attrs = append(attrs, semconv.ServerAddress(addr))
	}

	if port, ok := serverPort(req); ok {
		attrs = append(attrs, semconv.ServerPort(port))
	}

	e.semConvMetricRegistry.HTTPServerRequestDuration().Record(req.Context(), end.Sub(start).Seconds(),
		httpconv.RequestMethodAttr(req.Method), req.Header.Get("X-Forwarded-Proto"), attrs...)
}

func serverAddress(req *http.Request) string {
	host := req.Host
	if host == "" {
		return ""
	}

	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	}

	return host
}

func serverPort(req *http.Request) (int, bool) {
	if port, ok := portFromLocalAddr(req.Context()); ok {
		return port, true
	}

	if port, ok := explicitPortFromHost(req.Host); ok {
		return port, true
	}

	return defaultPortFromScheme(req)
}

func portFromLocalAddr(ctx context.Context) (int, bool) {
	if ctx == nil {
		return 0, false
	}

	val := ctx.Value(http.LocalAddrContextKey)
	if val == nil {
		return 0, false
	}

	switch addr := val.(type) {
	case *net.TCPAddr:
		if addr.Port > 0 {
			return addr.Port, true
		}
	case net.Addr:
		return parsePortFromString(addr.String())
	case fmt.Stringer:
		return parsePortFromString(addr.String())
	default:
		return parsePortFromString(fmt.Sprint(addr))
	}

	return 0, false
}

func explicitPortFromHost(host string) (int, bool) {
	if host == "" {
		return 0, false
	}

	if !strings.Contains(host, ":") {
		return 0, false
	}

	_, portStr, err := net.SplitHostPort(host)
	if err != nil {
		return 0, false
	}

	return parsePort(portStr)
}

func defaultPortFromScheme(req *http.Request) (int, bool) {
	switch scheme := derivedScheme(req); scheme {
	case "https":
		return 443, true
	case "http":
		return 80, true
	default:
		return 0, false
	}
}

func derivedScheme(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}

	if proto := parseForwardedProto(req.Header.Get("X-Forwarded-Proto")); proto != "" {
		return proto
	}

	if req.URL != nil && req.URL.Scheme != "" {
		return strings.ToLower(req.URL.Scheme)
	}

	return "http"
}

func parseForwardedProto(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(parts[0]))
}

func parsePortFromString(addr string) (int, bool) {
	if addr == "" {
		return 0, false
	}

	parsedAddr := addr
	if strings.HasPrefix(parsedAddr, ":") {
		parsedAddr = "127.0.0.1" + parsedAddr
	}

	_, portStr, err := net.SplitHostPort(parsedAddr)
	if err != nil {
		return 0, false
	}

	return parsePort(portStr)
}

func parsePort(portStr string) (int, bool) {
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return 0, false
	}

	return port, true
}
