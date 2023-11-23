package service

import (
	"net/http"

	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/trace"
)

type wrapper struct {
	rt http.RoundTripper
}

func (t *wrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	var span trace.Span
	if tracer := tracing.TracerFromContext(req.Context()); tracer != nil {
		tracingCtx := tracing.Propagator(req.Context(), req.Header)
		tracingCtx, span = tracer.Start(tracingCtx, "reverse-proxy", trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		req = req.WithContext(tracingCtx)

		tracing.LogClientRequest(span, req)
	}

	response, err := t.rt.RoundTrip(req)
	if err != nil {
		statusCode := computeStatusCode(err)
		tracing.LogResponseCode(span, statusCode, trace.SpanKindClient)
		return response, err
	}

	if span != nil {
		tracing.LogResponseCode(span, response.StatusCode, trace.SpanKindClient)
	}

	return response, nil
}

func newTracingRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &wrapper{rt: rt}
}
