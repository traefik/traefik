package tracing

import (
	"reflect"
	"testing"
)

func TestTracingNewForwarderMiddleware(t *testing.T) {
	trace := &Tracing{
		SpanNameLimit: 101,
	}

	testCases := []struct {
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

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			want := &forwarderMiddleware{
				Tracing:  trace,
				frontend: test.frontend,
				backend:  test.backend,
				opName:   test.name,
			}
			if got := test.tracer.NewForwarderMiddleware(test.frontend, test.backend); !reflect.DeepEqual(got, want) || len(want.opName) > trace.SpanNameLimit {
				t.Errorf("Tracing.NewForwarderMiddleware() = %+v, want %+v", got, want)
			}
		})
	}
}
