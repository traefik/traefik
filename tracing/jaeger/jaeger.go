package jaeger

import (
	"fmt"
	"io"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
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
	Gen128Bit              bool    `description:"generate 128 bit span IDs." export:"true"`
	Propagation            string  `description:"which propgation format to use (jaeger/b3)." export:"true"`
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

	opts := []jaegercfg.Option{
		jaegercfg.Logger(newJaegerLogger()),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.Gen128Bit(c.Gen128Bit),
	}

	switch c.Propagation {
	case "b3":
		p := zipkin.NewZipkinB3HTTPHeaderPropagator()
		opts = append(opts,
			jaegercfg.Injector(opentracing.HTTPHeaders, p),
			jaegercfg.Extractor(opentracing.HTTPHeaders, p),
		)
	case "jaeger", "":
	default:
		return nil, nil, fmt.Errorf("unknown propagation format: %s", c.Propagation)
	}

	// Initialize tracer with a logger and a metrics factory
	closer, err := jcfg.InitGlobalTracer(
		componentName,
		opts...,
	)
	if err != nil {
		log.WithoutContext().Warnf("Could not initialize jaeger tracer: %s", err.Error())
		return nil, nil, err
	}
	log.WithoutContext().Debug("Jaeger tracer configured")

	return opentracing.GlobalTracer(), closer, nil
}
