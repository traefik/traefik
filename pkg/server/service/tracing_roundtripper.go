package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type wrapper struct {
	semConvMetricRegistry *metrics.SemConvMetricsRegistry
	rt                    http.RoundTripper
}

func (t *wrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	var span trace.Span
	var tracingCtx context.Context
	if tracer := tracing.TracerFromContext(req.Context()); tracer != nil {
		tracingCtx, span = tracer.Start(req.Context(), "ReverseProxy", trace.WithSpanKind(trace.SpanKindClient))

		req = req.WithContext(tracingCtx)

		tracing.LogClientRequest(span, req)
		tracing.InjectContextIntoCarrier(req)
	}

	response, err := t.rt.RoundTrip(req)
	statusCode := response.StatusCode
	if err != nil {
		statusCode = computeStatusCode(err)
	}

	tracing.LogResponseCode(span, statusCode, trace.SpanKindClient)

	end := time.Now()

	if span != nil {
		span.End(trace.WithTimestamp(end))
	}

	if t.semConvMetricRegistry != nil && t.semConvMetricRegistry.HttpClientRequestDuration() != nil {
		var attrs []attribute.KeyValue

		if statusCode < 100 || statusCode >= 600 {
			attrs = append(attrs, attribute.Key("error.type").String(fmt.Sprintf("Invalid HTTP status code ; %d", statusCode)))
		} else if statusCode >= 400 {
			attrs = append(attrs, attribute.Key("error.type").Int(statusCode))
		}

		attrs = append(attrs, semconv.HTTPRequestMethodKey.String(req.Method))
		attrs = append(attrs, semconv.HTTPResponseStatusCode(statusCode))
		attrs = append(attrs, semconv.NetworkProtocolName(strings.ToLower(req.Proto)))
		attrs = append(attrs, semconv.NetworkProtocolVersion(proto(req.Proto)))
		attrs = append(attrs, semconv.ServerAddress(req.URL.Host))

		_, port, err := net.SplitHostPort(req.URL.Host)
		if err != nil {
			switch req.URL.Scheme {
			case "http":
				attrs = append(attrs, semconv.ServerPort(80))
			case "https":
				attrs = append(attrs, semconv.ServerPort(443))
			}
		} else {
			intPort, _ := strconv.Atoi(port)
			attrs = append(attrs, semconv.ServerPort(intPort))
		}

		attrs = append(attrs, semconv.URLScheme(req.Header.Get("X-Forwarded-Proto")))

		t.semConvMetricRegistry.HttpClientRequestDuration().Record(req.Context(), end.Sub(start).Seconds(), metric.WithAttributes(attrs...))
	}

	return response, nil
}

func newTracingRoundTripper(semConvMetricRegistry *metrics.SemConvMetricsRegistry, rt http.RoundTripper) http.RoundTripper {
	return &wrapper{
		semConvMetricRegistry: semConvMetricRegistry,
		rt:                    rt,
	}
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
