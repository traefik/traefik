package zipkin

import (
	"io"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

// Name sets the name of this tracer
const Name = "zipkin"

// Config provides configuration settings for a jaeger tracer
type Config struct {
	HTTPEndpoint string `description:"HTTP Endpoint to report traces to." export:"true"`
	SameSpan     bool   `description:"Use ZipKin SameSpan RPC stule traces." export:"true"`
	ID128Bit     bool   `description:"Use ZipKin 128 bit root span IDs." export:"true"`
	Debug        bool   `description:"Enable zipkin debug." export:"true"`
}

// Setup sets up the tracer
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	collector, err := zipkin.NewHTTPCollector("0.0.0.0:0")
	recorder := zipkin.NewRecorder(collector, c.Debug, c.HTTPEndpoint, serviceName)
	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(c.SameSpan),
		zipkin.TraceID128Bit(c.ID128Bit),
	)

	if err != nil {
		return nil, nil, err
	}

	return tracer, nil, nil
}
