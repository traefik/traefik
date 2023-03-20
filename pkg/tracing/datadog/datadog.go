package datadog

import (
	"io"
	"net"
	"os"
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
	LocalAgentHostPort string `description:"Sets the Datadog Agent host:port." json:"localAgentHostPort,omitempty" toml:"localAgentHostPort,omitempty" yaml:"localAgentHostPort,omitempty"`
	LocalAgentSocket   string `description:"Sets the socket for the Datadog Agent." json:"localAgentSocket,omitempty" toml:"localAgentSocket,omitempty" yaml:"localAgentSocket,omitempty"`
	// Deprecated: use GlobalTags instead.
	GlobalTag                  string            `description:"Sets a key:value tag on all spans." json:"globalTag,omitempty" toml:"globalTag,omitempty" yaml:"globalTag,omitempty" export:"true"`
	GlobalTags                 map[string]string `description:"Sets a list of key:value tags on all spans." json:"globalTags,omitempty" toml:"globalTags,omitempty" yaml:"globalTags,omitempty" export:"true"`
	Debug                      bool              `description:"Enables Datadog debug." json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty" export:"true"`
	PrioritySampling           bool              `description:"Enables priority sampling. When using distributed tracing, this option must be enabled in order to get all the parts of a distributed trace sampled." json:"prioritySampling,omitempty" toml:"prioritySampling,omitempty" yaml:"prioritySampling,omitempty" export:"true"`
	TraceIDHeaderName          string            `description:"Sets the header name used to store the trace ID." json:"traceIDHeaderName,omitempty" toml:"traceIDHeaderName,omitempty" yaml:"traceIDHeaderName,omitempty" export:"true"`
	ParentIDHeaderName         string            `description:"Sets the header name used to store the parent ID." json:"parentIDHeaderName,omitempty" toml:"parentIDHeaderName,omitempty" yaml:"parentIDHeaderName,omitempty" export:"true"`
	SamplingPriorityHeaderName string            `description:"Sets the header name used to store the sampling priority." json:"samplingPriorityHeaderName,omitempty" toml:"samplingPriorityHeaderName,omitempty" yaml:"samplingPriorityHeaderName,omitempty" export:"true"`
	BagagePrefixHeaderName     string            `description:"Sets the header name prefix used to store baggage items in a map." json:"bagagePrefixHeaderName,omitempty" toml:"bagagePrefixHeaderName,omitempty" yaml:"bagagePrefixHeaderName,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	host, ok := os.LookupEnv("DD_AGENT_HOST")
	if !ok {
		host = "localhost"
	}

	port, ok := os.LookupEnv("DD_TRACE_AGENT_PORT")
	if !ok {
		port = "8126"
	}

	c.LocalAgentHostPort = net.JoinHostPort(host, port)
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	opts := []datadog.StartOption{
		datadog.WithServiceName(serviceName),
		datadog.WithDebugMode(c.Debug),
		datadog.WithPropagator(datadog.NewPropagator(&datadog.PropagatorConfig{
			TraceHeader:    c.TraceIDHeaderName,
			ParentHeader:   c.ParentIDHeaderName,
			PriorityHeader: c.SamplingPriorityHeaderName,
			BaggagePrefix:  c.BagagePrefixHeaderName,
		})),
	}

	if c.LocalAgentSocket != "" {
		opts = append(opts, datadog.WithUDS(c.LocalAgentSocket))
	} else {
		opts = append(opts, datadog.WithAgentAddr(c.LocalAgentHostPort))
	}

	for k, v := range c.GlobalTags {
		opts = append(opts, datadog.WithGlobalTag(k, v))
	}

	if c.GlobalTag != "" {
		log.WithoutContext().Warn(`Datadog: option "globalTag" is deprecated, please use "globalTags" instead.`)

		key, value, _ := strings.Cut(c.GlobalTag, ":")

		// Don't override a tag already defined with the new option.
		if _, ok := c.GlobalTags[key]; !ok {
			opts = append(opts, datadog.WithGlobalTag(key, value))
		}
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
