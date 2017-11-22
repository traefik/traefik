package jaeger

import (
	"io"

	"github.com/containous/traefik/log"
	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	jaegermet "github.com/uber/jaeger-lib/metrics"
)

// Name sets the name of this tracer
const Name = "jaeger"

// Config provides configuration settings for a jaeger tracer
type Config struct {
	SamplingType  string  `description:"set the samplingn type." export:"true"`
	SamplingParam float64 `description:"set the sampling parameter." export:"true"`
}

// Setup sets up the tracer
func (c *Config) Setup(componentName string) (opentracing.Tracer, io.Closer, error) {
	jcfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  c.SamplingType,
			Param: c.SamplingParam,
		},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := jaegermet.NullFactory
	closer, err := jcfg.InitGlobalTracer(
		componentName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Warnf("Could not initialize jaeger tracer: %v", err)
		return nil, nil, err
	}
	log.Debugf("jaeger tracer configured", err)

	return opentracing.GlobalTracer(), closer, nil
}
