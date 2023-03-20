package metrics

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/containous/alice"
	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/middlewares/capture"
	"github.com/traefik/traefik/v2/pkg/middlewares/retry"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
)

const (
	protoHTTP      = "http"
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
	openConnsGauge       gokitmetrics.Gauge
	reqsBytesCounter     gokitmetrics.Counter
	respsBytesCounter    gokitmetrics.Counter
	baseLabels           []string
}

// NewEntryPointMiddleware creates a new metrics middleware for an Entrypoint.
func NewEntryPointMiddleware(ctx context.Context, next http.Handler, registry metrics.Registry, entryPointName string) http.Handler {
	log.FromContext(middlewares.GetLoggerCtx(ctx, nameEntrypoint, typeName)).Debug("Creating middleware")

	return &metricsMiddleware{
		next:                 next,
		reqsCounter:          registry.EntryPointReqsCounter(),
		reqsTLSCounter:       registry.EntryPointReqsTLSCounter(),
		reqDurationHistogram: registry.EntryPointReqDurationHistogram(),
		openConnsGauge:       registry.EntryPointOpenConnsGauge(),
		reqsBytesCounter:     registry.EntryPointReqsBytesCounter(),
		respsBytesCounter:    registry.EntryPointRespsBytesCounter(),
		baseLabels:           []string{"entrypoint", entryPointName},
	}
}

// NewRouterMiddleware creates a new metrics middleware for a Router.
func NewRouterMiddleware(ctx context.Context, next http.Handler, registry metrics.Registry, routerName string, serviceName string) http.Handler {
	log.FromContext(middlewares.GetLoggerCtx(ctx, nameRouter, typeName)).Debug("Creating middleware")

	return &metricsMiddleware{
		next:                 next,
		reqsCounter:          registry.RouterReqsCounter(),
		reqsTLSCounter:       registry.RouterReqsTLSCounter(),
		reqDurationHistogram: registry.RouterReqDurationHistogram(),
		openConnsGauge:       registry.RouterOpenConnsGauge(),
		reqsBytesCounter:     registry.RouterReqsBytesCounter(),
		respsBytesCounter:    registry.RouterRespsBytesCounter(),
		baseLabels:           []string{"router", routerName, "service", serviceName},
	}
}

// NewServiceMiddleware creates a new metrics middleware for a Service.
func NewServiceMiddleware(ctx context.Context, next http.Handler, registry metrics.Registry, serviceName string) http.Handler {
	log.FromContext(middlewares.GetLoggerCtx(ctx, nameService, typeName)).Debug("Creating middleware")

	return &metricsMiddleware{
		next:                 next,
		reqsCounter:          registry.ServiceReqsCounter(),
		reqsTLSCounter:       registry.ServiceReqsTLSCounter(),
		reqDurationHistogram: registry.ServiceReqDurationHistogram(),
		openConnsGauge:       registry.ServiceOpenConnsGauge(),
		reqsBytesCounter:     registry.ServiceReqsBytesCounter(),
		respsBytesCounter:    registry.ServiceRespsBytesCounter(),
		baseLabels:           []string{"service", serviceName},
	}
}

// WrapEntryPointHandler Wraps metrics entrypoint to alice.Constructor.
func WrapEntryPointHandler(ctx context.Context, registry metrics.Registry, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return NewEntryPointMiddleware(ctx, next, registry, entryPointName), nil
	}
}

// WrapRouterHandler Wraps metrics router to alice.Constructor.
func WrapRouterHandler(ctx context.Context, registry metrics.Registry, routerName string, serviceName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return NewRouterMiddleware(ctx, next, registry, routerName, serviceName), nil
	}
}

// WrapServiceHandler Wraps metrics service to alice.Constructor.
func WrapServiceHandler(ctx context.Context, registry metrics.Registry, serviceName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return NewServiceMiddleware(ctx, next, registry, serviceName), nil
	}
}

func (m *metricsMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var labels []string
	labels = append(labels, m.baseLabels...)
	labels = append(labels, "method", getMethod(req), "protocol", getRequestProtocol(req))

	openConnsGauge := m.openConnsGauge.With(labels...)
	openConnsGauge.Add(1)
	defer openConnsGauge.Add(-1)

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
		for i := 0; i < len(m.baseLabels); i += 2 {
			ctx = log.With(ctx, log.Str(m.baseLabels[i], m.baseLabels[i+1]))
		}
		log.FromContext(ctx).WithError(err).Errorf("Could not get Capture")
		return
	}

	next := m.next
	if capt.NeedsReset(rw) {
		next = capt.Reset(m.next)
	}

	start := time.Now()
	next.ServeHTTP(rw, req)

	labels = append(labels, "code", strconv.Itoa(capt.StatusCode()))
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

func containsHeader(req *http.Request, name, value string) bool {
	items := strings.Split(req.Header.Get(name), ",")
	for _, item := range items {
		if value == strings.ToLower(strings.TrimSpace(item)) {
			return true
		}
	}
	return false
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
		log.WithoutContext().Warnf("Invalid HTTP method encoding: %s", r.Method)
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
