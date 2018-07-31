package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracingNewForwarderMiddleware(t *testing.T) {
	testCases := []struct {
		desc     string
		tracer   *Tracing
		frontend string
		backend  string
		expected *forwarderMiddleware
	}{
		{
			desc: "Simple Forward Tracer without truncation and hashing",
			tracer: &Tracing{
				SpanNameLimit: 101,
			},
			frontend: "some-service.domain.tld",
			backend:  "some-service.domain.tld",
			expected: &forwarderMiddleware{
				Tracing: &Tracing{
					SpanNameLimit: 101,
				},
				frontend: "some-service.domain.tld",
				backend:  "some-service.domain.tld",
				opName:   "forward some-service.domain.tld/some-service.domain.tld",
			},
		}, {
			desc: "Simple Forward Tracer with truncation and hashing",
			tracer: &Tracing{
				SpanNameLimit: 101,
			},
			frontend: "some-service-100.slug.namespace.environment.domain.tld",
			backend:  "some-service-100.slug.namespace.environment.domain.tld",
			expected: &forwarderMiddleware{
				Tracing: &Tracing{
					SpanNameLimit: 101,
				},
				frontend: "some-service-100.slug.namespace.environment.domain.tld",
				backend:  "some-service-100.slug.namespace.environment.domain.tld",
				opName:   "forward some-service-100.slug.namespace.enviro.../some-service-100.slug.namespace.enviro.../bc4a0d48",
			},
		},
		{
			desc: "Exactly 101 chars",
			tracer: &Tracing{
				SpanNameLimit: 101,
			},
			frontend: "some-service1.namespace.environment.domain.tld",
			backend:  "some-service1.namespace.environment.domain.tld",
			expected: &forwarderMiddleware{
				Tracing: &Tracing{
					SpanNameLimit: 101,
				},
				frontend: "some-service1.namespace.environment.domain.tld",
				backend:  "some-service1.namespace.environment.domain.tld",
				opName:   "forward some-service1.namespace.environment.domain.tld/some-service1.namespace.environment.domain.tld",
			},
		},
		{
			desc: "More than 101 chars",
			tracer: &Tracing{
				SpanNameLimit: 101,
			},
			frontend: "some-service1.frontend.namespace.environment.domain.tld",
			backend:  "some-service1.backend.namespace.environment.domain.tld",
			expected: &forwarderMiddleware{
				Tracing: &Tracing{
					SpanNameLimit: 101,
				},
				frontend: "some-service1.frontend.namespace.environment.domain.tld",
				backend:  "some-service1.backend.namespace.environment.domain.tld",
				opName:   "forward some-service1.frontend.namespace.envir.../some-service1.backend.namespace.enviro.../fa49dd23",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.tracer.NewForwarderMiddleware(test.frontend, test.backend)

			assert.Equal(t, test.expected, actual)
			assert.True(t, len(test.expected.opName) <= test.tracer.SpanNameLimit)
		})
	}
}
