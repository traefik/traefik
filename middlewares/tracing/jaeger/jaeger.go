package jaeger

import (
	"io"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegermet "github.com/uber/jaeger-lib/metrics"
)

// Name sets the name of this tracer
const Name = "jaeger"

// Config provides configuration settings for a jaeger tracer
type Config struct {
	SamplingServerURL      string  `description:"set the sampling server url." export:"false"`
	SamplingType           string  `description:"set the sampling type." export:"true"`
	SamplingParam          float64 `description:"set the sampling parameter." export:"true"`
	LocalAgentHostPort     string  `description:"set jaeger-agent's host:port that the reporter will used." export:"false"`
	TraceContextHeaderName string  `description:"set the header to use for the trace-id." export:"true"`
}

// Setup sets up the tracer
func (c *Config) Setup(componentName string) (opentracing.Tracer, io.Closer, error) {
	jcfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			SamplingServerURL: c.SamplingServerURL,
			Type:              c.SamplingType,
			Param:             c.SamplingParam,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: c.LocalAgentHostPort,
		},
		Headers: &jaeger.HeadersConfig{
			TraceContextHeaderName: c.TraceContextHeaderName,
		},
	}

	jMetricsFactory := jaegermet.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := jcfg.InitGlobalTracer(
		componentName,
		jaegercfg.Logger(&jaegerLogger{}),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Warnf("Could not initialize jaeger tracer: %s", err.Error())
		return nil, nil, err
	}
	log.Debug("Jaeger tracer configured")

	return opentracing.GlobalTracer(), closer, nil
}
