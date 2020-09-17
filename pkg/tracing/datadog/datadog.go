package datadog

import (
	"io"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	datadog "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Name sets the name of this tracer.
const Name = "datadog"

// Config provides configuration settings for a datadog tracer.
type Config struct {
	LocalAgentHostPort         string `description:"Set datadog-agent's host:port that the reporter will used." json:"localAgentHostPort,omitempty" toml:"localAgentHostPort,omitempty" yaml:"localAgentHostPort,omitempty"`
	GlobalTag                  string `description:"Key:Value tag to be set on all the spans." json:"globalTag,omitempty" toml:"globalTag,omitempty" yaml:"globalTag,omitempty" export:"true"`
	Debug                      bool   `description:"Enable Datadog debug." json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty" export:"true"`
	PrioritySampling           bool   `description:"Enable priority sampling. When using distributed tracing, this option must be enabled in order to get all the parts of a distributed trace sampled." json:"prioritySampling,omitempty" toml:"prioritySampling,omitempty" yaml:"prioritySampling,omitempty"`
	TraceIDHeaderName          string `description:"Specifies the header name that will be used to store the trace ID." json:"traceIDHeaderName,omitempty" toml:"traceIDHeaderName,omitempty" yaml:"traceIDHeaderName,omitempty" export:"true"`
	ParentIDHeaderName         string `description:"Specifies the header name that will be used to store the parent ID." json:"parentIDHeaderName,omitempty" toml:"parentIDHeaderName,omitempty" yaml:"parentIDHeaderName,omitempty" export:"true"`
	SamplingPriorityHeaderName string `description:"Specifies the header name that will be used to store the sampling priority." json:"samplingPriorityHeaderName,omitempty" toml:"samplingPriorityHeaderName,omitempty" yaml:"samplingPriorityHeaderName,omitempty" export:"true"`
	BagagePrefixHeaderName     string `description:"Specifies the header name prefix that will be used to store baggage items in a map." json:"bagagePrefixHeaderName,omitempty" toml:"bagagePrefixHeaderName,omitempty" yaml:"bagagePrefixHeaderName,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.LocalAgentHostPort = "localhost:8126"
	c.GlobalTag = ""
	c.Debug = false
	c.PrioritySampling = false
}

// Setup sets up the tracer.
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
		datadog.WithPropagator(datadog.NewPropagator(&datadog.PropagatorConfig{
			TraceHeader:    c.TraceIDHeaderName,
			ParentHeader:   c.ParentIDHeaderName,
			PriorityHeader: c.SamplingPriorityHeaderName,
			BaggagePrefix:  c.BagagePrefixHeaderName,
		})),
	}
	if c.PrioritySampling {
		opts = append(opts, datadog.WithPrioritySampling())
	}
	tracer := ddtracer.New(opts...)

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("Datadog tracer configured")

	return tracer, nil, nil
}
