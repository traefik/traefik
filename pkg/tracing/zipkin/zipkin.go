package zipkin

import (
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Name sets the name of this tracer.
const Name = "zipkin"

// Config provides configuration settings for a zipkin tracer.
type Config struct {
	HTTPEndpoint string  `description:"Sets the HTTP Endpoint to report traces to." json:"httpEndpoint,omitempty" toml:"httpEndpoint,omitempty" yaml:"httpEndpoint,omitempty"`
	SameSpan     bool    `description:"Uses SameSpan RPC style traces." json:"sameSpan,omitempty" toml:"sameSpan,omitempty" yaml:"sameSpan,omitempty" export:"true"`
	ID128Bit     bool    `description:"Uses 128 bits root span IDs." json:"id128Bit,omitempty" toml:"id128Bit,omitempty" yaml:"id128Bit,omitempty" export:"true"`
	SampleRate   float64 `description:"Sets the rate between 0.0 and 1.0 of requests to trace." json:"sampleRate,omitempty" toml:"sampleRate,omitempty" yaml:"sampleRate,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.HTTPEndpoint = "http://localhost:9411/api/v2/spans"
	c.SameSpan = false
	c.ID128Bit = true
	c.SampleRate = 1.0
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	// create our local endpoint
	endpoint, err := zipkin.NewEndpoint(serviceName, "0.0.0.0:0")
	if err != nil {
		return nil, nil, err
	}

	// create our sampler
	sampler, err := zipkin.NewBoundarySampler(c.SampleRate, time.Now().Unix())
	if err != nil {
		return nil, nil, err
	}

	// create the span reporter
	reporter := http.NewReporter(c.HTTPEndpoint)

	// create the native Zipkin tracer
	nativeTracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSharedSpans(c.SameSpan),
		zipkin.WithTraceID128Bit(c.ID128Bit),
		zipkin.WithSampler(sampler),
	)
	if err != nil {
		return nil, nil, err
	}

	// wrap the Zipkin native tracer with the OpenTracing Bridge
	tracer := zipkinot.Wrap(nativeTracer)

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("Zipkin tracer configured")

	return tracer, reporter, nil
}
