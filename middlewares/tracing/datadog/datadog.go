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
}

// Setup sets up the tracer
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	tag := strings.SplitN(c.GlobalTag, ":", 2)

	value := ""
	if len(tag) == 2 {
		value = tag[1]
	}

	tracer := ddtracer.New(
		datadog.WithAgentAddr(c.LocalAgentHostPort),
		datadog.WithServiceName(serviceName),
		datadog.WithGlobalTag(tag[0], value),
		datadog.WithDebugMode(c.Debug),
	)

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.Debug("DataDog tracer configured")

	return tracer, nil, nil
}
