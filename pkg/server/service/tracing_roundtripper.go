package service

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/trace"
)

type wrapper struct {
	rt http.RoundTripper
}

func (t *wrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	var span trace.Span
	var tracer *tracing.Tracer
	if tracer = tracing.TracerFromContext(req.Context()); tracer != nil {
		var tracingCtx context.Context
		tracingCtx, span = tracer.Start(req.Context(), "ReverseProxy", trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		req = req.WithContext(tracingCtx)

		tracer.CaptureClientRequest(span, req)
		tracing.InjectContextIntoCarrier(req)
	}

	response, err := t.rt.RoundTrip(req)
	if err != nil {
		statusCode := computeStatusCode(err)
		tracer.CaptureResponse(span, nil, statusCode, trace.SpanKindClient)

		return response, err
	}

	tracer.CaptureResponse(span, response.Header, response.StatusCode, trace.SpanKindClient)

	return response, nil
}

func newTracingRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &wrapper{rt: rt}
}
