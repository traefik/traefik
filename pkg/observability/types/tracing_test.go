package types

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingVerbosity_Allows(t *testing.T) {
	tests := []struct {
		desc   string
		from   TracingVerbosity
		to     TracingVerbosity
		allows bool
	}{
		{
			desc:   "minimal vs minimal",
			from:   MinimalVerbosity,
			to:     MinimalVerbosity,
			allows: true,
		},
		{
			desc:   "minimal vs detailed",
			from:   MinimalVerbosity,
			to:     DetailedVerbosity,
			allows: false,
		},
		{
			desc:   "detailed vs minimal",
			from:   DetailedVerbosity,
			to:     MinimalVerbosity,
			allows: true,
		},
		{
			desc:   "detailed vs detailed",
			from:   DetailedVerbosity,
			to:     DetailedVerbosity,
			allows: true,
		},
		{
			desc:   "unknown vs minimal",
			from:   TracingVerbosity("unknown"),
			to:     MinimalVerbosity,
			allows: true,
		},
		{
			desc:   "unknown vs detailed",
			from:   TracingVerbosity("unknown"),
			to:     DetailedVerbosity,
			allows: false,
		},
		{
			desc:   "minimal vs unknown",
			from:   MinimalVerbosity,
			to:     TracingVerbosity("unknown"),
			allows: false,
		},
		{
			desc:   "detailed vs unknown",
			from:   DetailedVerbosity,
			to:     TracingVerbosity("unknown"),
			allows: false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.allows, test.from.Allows(test.to))
		})
	}
}

func TestOTelTracing_SamplerFromEnv(t *testing.T) {
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(collector.Close)

	tests := []struct {
		desc       string
		env        map[string]string
		sampleRate float64
		sampled    bool
	}{
		{
			desc:       "sampleRate applies without OTEL_TRACES_SAMPLER",
			sampleRate: 0,
			sampled:    false,
		},
		{
			desc:       "sampleRate applies with a blank OTEL_TRACES_SAMPLER",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": " "},
			sampleRate: 0,
			sampled:    false,
		},
		{
			desc:       "sampleRate applies with an unsupported OTEL_TRACES_SAMPLER",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "jaeger_remote"},
			sampleRate: 0,
			sampled:    false,
		},
		{
			desc:       "always_off overrides sampleRate",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "always_off"},
			sampleRate: 1,
			sampled:    false,
		},
		{
			desc:       "always_on in another case and with spaces overrides sampleRate",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": " ALWAYS_ON "},
			sampleRate: 0,
			sampled:    true,
		},
		{
			desc:       "always_on ignores OTEL_TRACES_SAMPLER_ARG",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "always_on", "OTEL_TRACES_SAMPLER_ARG": "10%"},
			sampleRate: 0,
			sampled:    true,
		},
		{
			desc:       "parentbased_always_on overrides sampleRate",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "parentbased_always_on"},
			sampleRate: 0,
			sampled:    true,
		},
		{
			desc:       "parentbased_always_off overrides sampleRate",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "parentbased_always_off"},
			sampleRate: 1,
			sampled:    false,
		},
		{
			desc:       "traceidratio without OTEL_TRACES_SAMPLER_ARG overrides sampleRate",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "traceidratio"},
			sampleRate: 0,
			sampled:    true,
		},
		{
			desc:       "traceidratio accepts an OTEL_TRACES_SAMPLER_ARG of 1",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "traceidratio", "OTEL_TRACES_SAMPLER_ARG": "1"},
			sampleRate: 0,
			sampled:    true,
		},
		{
			desc:       "traceidratio uses OTEL_TRACES_SAMPLER_ARG",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "traceidratio", "OTEL_TRACES_SAMPLER_ARG": "0"},
			sampleRate: 1,
			sampled:    false,
		},
		{
			desc:       "traceidratio ignores an empty OTEL_TRACES_SAMPLER_ARG",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "traceidratio", "OTEL_TRACES_SAMPLER_ARG": ""},
			sampleRate: 0,
			sampled:    true,
		},
		{
			desc:       "parentbased_traceidratio ignores a blank OTEL_TRACES_SAMPLER_ARG",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "parentbased_traceidratio", "OTEL_TRACES_SAMPLER_ARG": " "},
			sampleRate: 0,
			sampled:    true,
		},
		{
			desc:       "sampleRate applies with an invalid OTEL_TRACES_SAMPLER_ARG",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "parentbased_traceidratio", "OTEL_TRACES_SAMPLER_ARG": "10%"},
			sampleRate: 0,
			sampled:    false,
		},
		{
			desc:       "sampleRate applies with an OTEL_TRACES_SAMPLER_ARG above 1",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "traceidratio", "OTEL_TRACES_SAMPLER_ARG": "1.5"},
			sampleRate: 0,
			sampled:    false,
		},
		{
			desc:       "sampleRate applies with a negative OTEL_TRACES_SAMPLER_ARG",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "parentbased_traceidratio", "OTEL_TRACES_SAMPLER_ARG": "-0.1"},
			sampleRate: 0,
			sampled:    false,
		},
		{
			desc:       "sampleRate applies with a NaN OTEL_TRACES_SAMPLER_ARG",
			env:        map[string]string{"OTEL_TRACES_SAMPLER": "traceidratio", "OTEL_TRACES_SAMPLER_ARG": "NaN"},
			sampleRate: 0,
			sampled:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			for _, key := range []string{"OTEL_TRACES_SAMPLER", "OTEL_TRACES_SAMPLER_ARG"} {
				value, ok := test.env[key]
				t.Setenv(key, value)
				if !ok {
					require.NoError(t, os.Unsetenv(key))
				}
			}

			config := &OTelTracing{}
			config.SetDefaults()
			config.HTTP.Endpoint = collector.URL

			tracer, closer, err := config.Setup(t.Context(), "traefik", test.sampleRate, nil)
			require.NoError(t, err)
			t.Cleanup(func() { _ = closer.Close() })

			_, span := tracer.Start(t.Context(), "test")
			span.End()

			assert.Equal(t, test.sampled, span.SpanContext().IsSampled())
		})
	}
}
