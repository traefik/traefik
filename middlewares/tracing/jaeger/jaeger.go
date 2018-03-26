package jaeger

import (
	"io"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	jaegermet "github.com/uber/jaeger-lib/metrics"
)

// Name sets the name of this tracer
const Name = "jaeger"

// Config provides configuration settings for a jaeger tracer
type Config struct {
	SamplingServerURL  string  `description:"set the sampling server url." export:"false"`
	SamplingType       string  `description:"set the sampling type." export:"true"`
	SamplingParam      float64 `description:"set the sampling parameter." export:"true"`
	LocalAgentHostPort string  `description:"set jaeger-agent's host:port that the reporter will used." export:"false"`
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
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := jaegermet.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := jcfg.InitGlobalTracer(
		componentName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Warnf("Could not initialize jaeger tracer: %s", err.Error())
		return nil, nil, err
	}
	log.Debugf("jaeger tracer configured", err)

	return opentracing.GlobalTracer(), closer, nil
}
