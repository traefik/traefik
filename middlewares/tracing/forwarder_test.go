package tracing

import (
	"reflect"
	"testing"
)

func TestTracing_NewForwarderMiddleware(t *testing.T) {
	trace := &Tracing{
		SpanNameLimit: 101,
	}
	tests := []struct {
		desc     string
		frontend string
		backend  string
		name     string
		tracer   *Tracing
	}{
		{
			desc:     "Simple Forward Tracer with truncation and hashing",
			frontend: "some-service-100.slug.namespace.environment.domain.tld",
			backend:  "some-service-100.slug.namespace.environment.domain.tld",
			name:     "forward some-service-100.slug.namespace.enviro.../some-service-100.slug.namespace.enviro.../bc4a0d48",
			tracer:   trace,
		},
		{
			desc:     "Simple Forward Tracer without truncation and hashing",
			frontend: "some-service.domain.tld",
			backend:  "some-service.domain.tld",
			name:     "forward some-service.domain.tld/some-service.domain.tld",
			tracer:   trace,
		},
		{
			desc:     "Exactly 101 chars",
			frontend: "some-service1.namespace.environment.domain.tld",
			backend:  "some-service1.namespace.environment.domain.tld",
			name:     "forward some-service1.namespace.environment.domain.tld/some-service1.namespace.environment.domain.tld",
			tracer:   trace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			want := &forwarderMiddleware{
				Tracing:  trace,
				frontend: tt.frontend,
				backend:  tt.backend,
				opName:   tt.name,
			}
			if got := tt.tracer.NewForwarderMiddleware(tt.frontend, tt.backend); !reflect.DeepEqual(got, want) || len(want.opName) > trace.SpanNameLimit {
				t.Errorf("Tracing.NewForwarderMiddleware() = %+v, want %+v", got, want)
			}
		})
	}
}
