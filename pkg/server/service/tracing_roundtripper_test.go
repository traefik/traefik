package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

func TestTracingRoundTripper(t *testing.T) {
	type expected struct {
		name       string
		attributes []attribute.KeyValue
	}

	testCases := []struct {
		desc     string
		expected []expected
	}{
		{
			desc: "basic test",
			expected: []expected{
				{
					name: "initial",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "unspecified"),
					},
				},
				{
					name: "ReverseProxy",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "client"),
						attribute.String("http.request.method", "GET"),
						attribute.String("network.protocol.version", "1.1"),
						attribute.String("url.full", "http://www.test.com/search?q=Opentelemetry"),
						attribute.String("url.scheme", "http"),
						attribute.String("user_agent.original", "reverse-test"),
						attribute.String("network.peer.address", ""),
						attribute.String("server.address", "www.test.com"),
						attribute.String("network.peer.port", "80"),
						attribute.Int64("server.port", int64(80)),
						attribute.StringSlice("http.request.header.x-foo", []string{"foo", "bar"}),
						attribute.Int64("http.response.status_code", int64(404)),
						attribute.StringSlice("http.response.header.x-bar", []string{"foo", "bar"}),
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/search?q=Opentelemetry", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "reverse-test")
			req.Header.Set("X-Forwarded-Proto", "http")
			req.Header.Set("X-Foo", "foo")
			req.Header.Add("X-Foo", "bar")

			mockTracer := &mockTracer{}
			tracer := tracing.NewTracer(mockTracer, []string{"X-Foo"}, []string{"X-Bar"})
			initialCtx, initialSpan := tracer.Start(req.Context(), "initial")
			defer initialSpan.End()
			req = req.WithContext(initialCtx)

			tracingRoundTripper := newTracingRoundTripper(roundTripperFn(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Header: map[string][]string{
						"X-Bar": {"foo", "bar"},
					},
					StatusCode: http.StatusNotFound,
				}, nil
			}))

			_, err := tracingRoundTripper.RoundTrip(req)
			require.NoError(t, err)

			for i, span := range mockTracer.spans {
				assert.Equal(t, test.expected[i].name, span.name)
				assert.Equal(t, test.expected[i].attributes, span.attributes)
			}
		})
	}
}

type mockTracer struct {
	embedded.Tracer

	spans []*mockSpan
}

var _ trace.Tracer = &mockTracer{}

func (t *mockTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	config := trace.NewSpanStartConfig(opts...)
	span := &mockSpan{}
	span.SetName(name)
	span.SetAttributes(attribute.String("span.kind", config.SpanKind().String()))
	span.SetAttributes(config.Attributes()...)
	t.spans = append(t.spans, span)
	return trace.ContextWithSpan(ctx, span), span
}

// mockSpan is an implementation of Span that preforms no operations.
type mockSpan struct {
	embedded.Span

	name       string
	attributes []attribute.KeyValue
}

var _ trace.Span = &mockSpan{}

func (*mockSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}})
}
func (*mockSpan) IsRecording() bool                  { return false }
func (s *mockSpan) SetStatus(_ codes.Code, _ string) {}
func (s *mockSpan) SetAttributes(kv ...attribute.KeyValue) {
	s.attributes = append(s.attributes, kv...)
}
func (s *mockSpan) End(...trace.SpanEndOption)                  {}
func (s *mockSpan) RecordError(_ error, _ ...trace.EventOption) {}
func (s *mockSpan) AddEvent(_ string, _ ...trace.EventOption)   {}

func (s *mockSpan) SetName(name string) { s.name = name }

func (s *mockSpan) TracerProvider() trace.TracerProvider {
	return nil
}
