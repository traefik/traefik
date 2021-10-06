package zipkin

import (
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/traefik/traefik/log"
)

// Name sets the name of this tracer.
const Name = "zipkin"

// Config provides configuration settings for a zipkin tracer.
type Config struct {
	HTTPEndpoint string  `description:"HTTP Endpoint to report traces to." export:"false"`
	SameSpan     bool    `description:"Use ZipKin SameSpan RPC style traces." export:"true"`
	ID128Bit     bool    `description:"Use ZipKin 128 bit root span IDs." export:"true"`
	Debug        bool    `description:"Enable Zipkin debug." export:"true"`
	SampleRate   float64 `description:"Sets the rate between 0.0 and 1.0 of requests to trace." export:"true"`
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

	log.Debug("Zipkin tracer configured")

	return tracer, reporter, nil
}
