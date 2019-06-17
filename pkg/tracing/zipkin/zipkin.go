package zipkin

import (
	"io"
	"time"

	"github.com/containous/traefik/pkg/log"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
)

// Name sets the name of this tracer.
const Name = "zipkin"

// Config provides configuration settings for a zipkin tracer.
type Config struct {
	HTTPEndpoint string  `description:"HTTP Endpoint to report traces to." export:"false"`
	SameSpan     bool    `description:"Use Zipkin SameSpan RPC style traces." export:"true"`
	ID128Bit     bool    `description:"Use Zipkin 128 bit root span IDs." export:"true"`
	Debug        bool    `description:"Enable Zipkin debug." export:"true"`
	SampleRate   float64 `description:"The rate between 0.0 and 1.0 of requests to trace." export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.HTTPEndpoint = "http://localhost:9411/api/v1/spans"
	c.SameSpan = false
	c.ID128Bit = true
	c.Debug = false
	c.SampleRate = 1.0
}

// Setup sets up the tracer
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	collector, err := zipkin.NewHTTPCollector(c.HTTPEndpoint)
	if err != nil {
		return nil, nil, err
	}

	recorder := zipkin.NewRecorder(collector, c.Debug, "0.0.0.0:0", serviceName)

	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(c.SameSpan),
		zipkin.TraceID128Bit(c.ID128Bit),
		zipkin.DebugMode(c.Debug),
		zipkin.WithSampler(zipkin.NewBoundarySampler(c.SampleRate, time.Now().Unix())),
	)
	if err != nil {
		return nil, nil, err
	}

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("Zipkin tracer configured")

	return tracer, collector, nil
}
