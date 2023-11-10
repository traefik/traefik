package datadog

import (
	"io"
	"net"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	ddotel "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry"
	tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Name sets the name of this tracer.
const Name = "datadog"

// Config provides configuration settings for a datadog tracer.
type Config struct {
	LocalAgentHostPort         string            `description:"Sets the Datadog Agent host:port." json:"localAgentHostPort,omitempty" toml:"localAgentHostPort,omitempty" yaml:"localAgentHostPort,omitempty"`
	LocalAgentSocket           string            `description:"Sets the socket for the Datadog Agent." json:"localAgentSocket,omitempty" toml:"localAgentSocket,omitempty" yaml:"localAgentSocket,omitempty"`
	GlobalTags                 map[string]string `description:"Sets a list of key:value tags on all spans." json:"globalTags,omitempty" toml:"globalTags,omitempty" yaml:"globalTags,omitempty" export:"true"`
	Debug                      bool              `description:"Enables Datadog debug." json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty" export:"true"`
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
func (c *Config) Setup(serviceName string) (trace.Tracer, io.Closer, error) {
	logger := log.With().Str(logs.TracingProviderName, Name).Logger()
	opts := []tracer.StartOption{
		tracer.WithServiceName("traefik"),
		tracer.WithDebugMode(c.Debug),
		tracer.WithPropagator(tracer.NewPropagator(&tracer.PropagatorConfig{
			TraceHeader:    c.TraceIDHeaderName,
			ParentHeader:   c.ParentIDHeaderName,
			PriorityHeader: c.SamplingPriorityHeaderName,
			BaggagePrefix:  c.BagagePrefixHeaderName,
		})),
		tracer.WithLogger(logs.NewDatadogLogger(logger)),
	}

	if c.LocalAgentSocket != "" {
		opts = append(opts, tracer.WithUDS(c.LocalAgentSocket))
	} else {
		opts = append(opts, tracer.WithAgentAddr(c.LocalAgentHostPort))
	}

	for k, v := range c.GlobalTags {
		opts = append(opts, tracer.WithGlobalTag(k, v))
	}

	tracerProvider := ddotel.NewTracerProvider(opts...)
	otel.SetTracerProvider(tracerProvider)

	logger.Debug().Msg("Datadog tracer configured")

	return tracerProvider.Tracer(serviceName, trace.WithInstrumentationVersion(version.Version)), tpCloser{provider: tracerProvider}, nil
}

// tpCloser converts a TraceProvider into an io.Closer.
type tpCloser struct {
	provider *ddotel.TracerProvider
}

func (t tpCloser) Close() error {
	return t.provider.Shutdown()
}
