package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/semconv/v1.37.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

type wrapper struct {
	semConvMetricRegistry *metrics.SemConvMetricsRegistry
	rt                    http.RoundTripper
}

func newObservabilityRoundTripper(semConvMetricRegistry *metrics.SemConvMetricsRegistry, rt http.RoundTripper) http.RoundTripper {
	return &wrapper{
		semConvMetricRegistry: semConvMetricRegistry,
		rt:                    rt,
	}
}

func (t *wrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	var span trace.Span
	var tracingCtx context.Context
	var tracer *tracing.Tracer
	if tracer = tracing.TracerFromContext(req.Context()); tracer != nil && observability.TracingEnabled(req.Context()) {
		tracingCtx, span = tracer.Start(req.Context(), "ReverseProxy", trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		req = req.WithContext(tracingCtx)

		tracer.CaptureClientRequest(span, req)
		tracing.InjectContextIntoCarrier(req)
	}

	var statusCode int
	var headers http.Header
	response, err := t.rt.RoundTrip(req)
	if err != nil {
		statusCode = ComputeStatusCode(err)
	}
	if response != nil {
		statusCode = response.StatusCode
		headers = response.Header
	}

	if tracer != nil {
		tracer.CaptureResponse(span, headers, statusCode, trace.SpanKindClient)
	}

	end := time.Now()

	// Ending the span as soon as the response is handled because we want to use the same end time for the trace and the metric.
	// If any errors happen earlier, this span will be close by the defer instruction.
	if span != nil {
		span.End(trace.WithTimestamp(end))
	}

	if !observability.SemConvMetricsEnabled(req.Context()) ||
		t.semConvMetricRegistry == nil {
		return response, err
	}

	var attrs []attribute.KeyValue

	if statusCode < 100 || statusCode >= 600 {
		attrs = append(attrs, attribute.Key("error.type").String(fmt.Sprintf("Invalid HTTP status code %d", statusCode)))
	} else if statusCode >= 400 {
		attrs = append(attrs, attribute.Key("error.type").String(strconv.Itoa(statusCode)))
	}

	attrs = append(attrs, semconv.HTTPRequestMethodKey.String(req.Method))
	attrs = append(attrs, semconv.HTTPResponseStatusCode(statusCode))
	attrs = append(attrs, semconv.NetworkProtocolName(strings.ToLower(req.Proto)))
	attrs = append(attrs, semconv.NetworkProtocolVersion(observability.Proto(req.Proto)))

	_, port, splitErr := net.SplitHostPort(req.URL.Host)
	var serverPort int
	if splitErr != nil {
		switch req.URL.Scheme {
		case "http":
			serverPort = 80
			attrs = append(attrs, semconv.ServerPort(80))
		case "https":
			serverPort = 443
			attrs = append(attrs, semconv.ServerPort(443))
		}
	} else {
		intPort, _ := strconv.Atoi(port)
		serverPort = intPort
		attrs = append(attrs, semconv.ServerPort(intPort))
	}

	attrs = append(attrs, semconv.URLScheme(req.Header.Get("X-Forwarded-Proto")))

	// Convert method to httpconv enum
	var methodAttr httpconv.RequestMethodAttr
	switch req.Method {
	case http.MethodGet:
		methodAttr = httpconv.RequestMethodGet
	case http.MethodPost:
		methodAttr = httpconv.RequestMethodPost
	case http.MethodPut:
		methodAttr = httpconv.RequestMethodPut
	case http.MethodDelete:
		methodAttr = httpconv.RequestMethodDelete
	case http.MethodHead:
		methodAttr = httpconv.RequestMethodHead
	case http.MethodOptions:
		methodAttr = httpconv.RequestMethodOptions
	case http.MethodPatch:
		methodAttr = httpconv.RequestMethodPatch
	default:
		methodAttr = httpconv.RequestMethodOther
	}

	t.semConvMetricRegistry.HTTPClientRequestDuration().Record(req.Context(), end.Sub(start).Seconds(),
		methodAttr, req.URL.Host, serverPort, attrs...)

	return response, err
}
