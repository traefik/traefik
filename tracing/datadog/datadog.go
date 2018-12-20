package datadog

import (
	"io"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	datadog "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Name sets the name of this tracer
const Name = "datadog"

// Config provides configuration settings for a datadog tracer
type Config struct {
	LocalAgentHostPort string `description:"Set datadog-agent's host:port that the reporter will used. Defaults to localhost:8126" export:"false"`
	GlobalTag          string `description:"Key:Value tag to be set on all the spans." export:"true"`
	Debug              bool   `description:"Enable DataDog debug." export:"true"`
	PrioritySampling   bool   `description:"Enable priority sampling. When using distributed tracing, this option must be enabled in order to get all the parts of a distributed trace sampled."`
}

// Setup sets up the tracer
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	tag := strings.SplitN(c.GlobalTag, ":", 2)

	value := ""
	if len(tag) == 2 {
		value = tag[1]
	}

	opts := []datadog.StartOption{
		datadog.WithAgentAddr(c.LocalAgentHostPort),
		datadog.WithServiceName(serviceName),
		datadog.WithGlobalTag(tag[0], value),
		datadog.WithDebugMode(c.Debug),
	}
	if c.PrioritySampling {
		opts = append(opts, datadog.WithPrioritySampling())
	}
	tracer := ddtracer.New(opts...)

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("DataDog tracer configured")

	return tracer, nil, nil
}
