package metrics

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/containous/alice"
	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/middlewares/retry"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"google.golang.org/grpc/codes"
)

const (
	protoHTTP      = "http"
	protoGRPC      = "grpc"
	protoGRPCWeb   = "grpc-web"
	protoSSE       = "sse"
	protoWebsocket = "websocket"
	typeName       = "Metrics"
	nameEntrypoint = "metrics-entrypoint"
	nameRouter     = "metrics-router"
	nameService    = "metrics-service"
)

type metricsMiddleware struct {
	next                 http.Handler
	reqsCounter          metrics.CounterWithHeaders
	reqsTLSCounter       gokitmetrics.Counter
	reqDurationHistogram metrics.ScalableHistogram
	reqsBytesCounter     gokitmetrics.Counter
	respsBytesCounter    gokitmetrics.Counter
	baseLabels           []string
	name                 string
}

// NewEntryPointMiddleware creates a new metrics middleware for an Entrypoint.
func NewEntryPointMiddleware(ctx context.Context, next http.Handler, registry metrics.Registry, entryPointName string) http.Handler {
	middlewares.GetLogger(ctx, nameEntrypoint, typeName).Debug().Msg("Creating middleware")

	return &metricsMiddleware{
		next:                 next,
		reqsCounter:          registry.EntryPointReqsCounter(),
		reqsTLSCounter:       registry.EntryPointReqsTLSCounter(),
		reqDurationHistogram: registry.EntryPointReqDurationHistogram(),
		reqsBytesCounter:     registry.EntryPointReqsBytesCounter(),
		respsBytesCounter:    registry.EntryPointRespsBytesCounter(),
		baseLabels:           []string{"entrypoint", entryPointName},
		name:                 nameEntrypoint,
	}
}

// NewRouterMiddleware creates a new metrics middleware for a Router.
func NewRouterMiddleware(ctx context.Context, next http.Handler, registry metrics.Registry, routerName string, serviceName string) http.Handler {
	middlewares.GetLogger(ctx, nameRouter, typeName).Debug().Msg("Creating middleware")

	return &metricsMiddleware{
		next:                 next,
		reqsCounter:          registry.RouterReqsCounter(),
		reqsTLSCounter:       registry.RouterReqsTLSCounter(),
		reqDurationHistogram: registry.RouterReqDurationHistogram(),
		reqsBytesCounter:     registry.RouterReqsBytesCounter(),
		respsBytesCounter:    registry.RouterRespsBytesCounter(),
		baseLabels:           []string{"router", routerName, "service", serviceName},
		name:                 nameRouter,
	}
}

// NewServiceMiddleware creates a new metrics middleware for a Service.
func NewServiceMiddleware(ctx context.Context, next http.Handler, registry metrics.Registry, serviceName string) http.Handler {
	middlewares.GetLogger(ctx, nameService, typeName).Debug().Msg("Creating middleware")

	return &metricsMiddleware{
		next:                 next,
		reqsCounter:          registry.ServiceReqsCounter(),
		reqsTLSCounter:       registry.ServiceReqsTLSCounter(),
		reqDurationHistogram: registry.ServiceReqDurationHistogram(),
		reqsBytesCounter:     registry.ServiceReqsBytesCounter(),
		respsBytesCounter:    registry.ServiceRespsBytesCounter(),
		baseLabels:           []string{"service", serviceName},
		name:                 nameService,
	}
}

// EntryPointMetricsHandler returns the metrics entrypoint handler.
func EntryPointMetricsHandler(ctx context.Context, registry metrics.Registry, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if registry == nil || !registry.IsEpEnabled() {
			return next, nil
		}

		return NewEntryPointMiddleware(ctx, next, registry, entryPointName), nil
	}
}

// RouterMetricsHandler returns the metrics router handler.
func RouterMetricsHandler(ctx context.Context, registry metrics.Registry, routerName string, serviceName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if registry == nil || !registry.IsRouterEnabled() {
			return next, nil
		}

		return NewRouterMiddleware(ctx, next, registry, routerName, serviceName), nil
	}
}

// ServiceMetricsHandler returns the metrics service handler.
func ServiceMetricsHandler(ctx context.Context, registry metrics.Registry, serviceName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if registry == nil || !registry.IsSvcEnabled() {
			return next, nil
		}

		return NewServiceMiddleware(ctx, next, registry, serviceName), nil
	}
}

func (m *metricsMiddleware) GetTracingInformation() (string, string) {
	return m.name, typeName
}

func (m *metricsMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !observability.MetricsEnabled(req.Context()) {
		m.next.ServeHTTP(rw, req)
		return
	}

	proto := getRequestProtocol(req)

	var labels []string
	labels = append(labels, m.baseLabels...)
	labels = append(labels, "method", getMethod(req))
	labels = append(labels, "protocol", proto)

	// TLS metrics
	if req.TLS != nil {
		var tlsLabels []string
		tlsLabels = append(tlsLabels, m.baseLabels...)
		tlsLabels = append(tlsLabels, "tls_version", traefiktls.GetVersion(req.TLS), "tls_cipher", traefiktls.GetCipherName(req.TLS))

		m.reqsTLSCounter.With(tlsLabels...).Add(1)
	}

	ctx := req.Context()

	capt, err := capture.FromContext(ctx)
	if err != nil {
		with := log.Ctx(ctx).With()
		for i := 0; i < len(m.baseLabels); i += 2 {
			with = with.Str(m.baseLabels[i], m.baseLabels[i+1])
		}
		logger := with.Logger()
		logger.Error().Err(err).Msg("Could not get Capture")
		observability.SetStatusErrorf(req.Context(), "Could not get Capture")
		return
	}

	next := m.next
	if capt.NeedsReset(rw) {
		next = capt.Reset(m.next)
	}

	start := time.Now()
	next.ServeHTTP(rw, req)

	code := capt.StatusCode()
	if proto == protoGRPC || proto == protoGRPCWeb {
		code = grpcStatusCode(rw)
	}

	labels = append(labels, "code", strconv.Itoa(code))
	m.reqDurationHistogram.With(labels...).ObserveFromStart(start)
	m.reqsCounter.With(req.Header, labels...).Add(1)
	m.respsBytesCounter.With(labels...).Add(float64(capt.ResponseSize()))
	m.reqsBytesCounter.With(labels...).Add(float64(capt.RequestSize()))
}

func getRequestProtocol(req *http.Request) string {
	switch {
	case isWebsocketRequest(req):
		return protoWebsocket
	case isSSERequest(req):
		return protoSSE
	case isGRPCWebRequest(req):
		return protoGRPCWeb
	case isGRPCRequest(req):
		return protoGRPC
	default:
		return protoHTTP
	}
}

// isWebsocketRequest determines if the specified HTTP request is a websocket handshake request.
func isWebsocketRequest(req *http.Request) bool {
	return containsHeader(req, "Connection", "upgrade") && containsHeader(req, "Upgrade", "websocket")
}

// isSSERequest determines if the specified HTTP request is a request for an event subscription.
func isSSERequest(req *http.Request) bool {
	return containsHeader(req, "Accept", "text/event-stream")
}

// isGRPCWebRequest determines if the specified HTTP request is a gRPC-Web request.
func isGRPCWebRequest(req *http.Request) bool {
	return strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc-web")
}

// isGRPCRequest determines if the specified HTTP request is a gRPC request.
func isGRPCRequest(req *http.Request) bool {
	return strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc")
}

// grpcStatusCode parses and returns the gRPC status code from the Grpc-Status header.
func grpcStatusCode(rw http.ResponseWriter) int {
	code := codes.Unknown
	if status := rw.Header().Get("Grpc-Status"); status != "" {
		if err := code.UnmarshalJSON([]byte(status)); err != nil {
			return int(code)
		}
	}
	return int(code)
}

func containsHeader(req *http.Request, name, value string) bool {
	items := strings.Split(req.Header.Get(name), ",")

	return slices.ContainsFunc(items, func(item string) bool {
		return value == strings.ToLower(strings.TrimSpace(item))
	})
}

// getMethod returns the request's method.
// It checks whether the method is a valid UTF-8 string.
// To restrict the (potentially infinite) number of accepted values for the method,
// and avoid unbounded memory issues,
// values that are not part of the set of HTTP verbs are replaced with EXTENSION_METHOD.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
// https://datatracker.ietf.org/doc/html/rfc2616/#section-5.1.1.
//
//nolint:usestdlibvars
func getMethod(r *http.Request) string {
	if !utf8.ValidString(r.Method) {
		log.Warn().Msgf("Invalid HTTP method encoding: %s", r.Method)
		return "NON_UTF8_HTTP_METHOD"
	}

	switch r.Method {
	case "HEAD", "GET", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", // https://datatracker.ietf.org/doc/html/rfc7231#section-4
		"PATCH", // https://datatracker.ietf.org/doc/html/rfc5789#section-2
		"PRI":   // https://datatracker.ietf.org/doc/html/rfc7540#section-11.6
		return r.Method
	default:
		return "EXTENSION_METHOD"
	}
}

type retryMetrics interface {
	ServiceRetriesCounter() gokitmetrics.Counter
}

// NewRetryListener instantiates a MetricsRetryListener with the given retryMetrics.
func NewRetryListener(retryMetrics retryMetrics, serviceName string) retry.Listener {
	return &RetryListener{retryMetrics: retryMetrics, serviceName: serviceName}
}

// RetryListener is an implementation of the RetryListener interface to
// record RequestMetrics about retry attempts.
type RetryListener struct {
	retryMetrics retryMetrics
	serviceName  string
}

// Retried tracks the retry in the RequestMetrics implementation.
func (m *RetryListener) Retried(_ *http.Request, _ int) {
	m.retryMetrics.ServiceRetriesCounter().With("service", m.serviceName).Add(1)
}
