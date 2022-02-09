package haystack

import (
	"io"
	"strings"
	"time"

	"github.com/ExpediaDotCom/haystack-client-go"
	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Name sets the name of this tracer.
const Name = "haystack"

// Config provides configuration settings for a haystack tracer.
type Config struct {
	LocalAgentHost          string `description:"Sets the Haystack Agent host." json:"localAgentHost,omitempty" toml:"localAgentHost,omitempty" yaml:"localAgentHost,omitempty"`
	LocalAgentPort          int    `description:"Sets the Haystack Agent port." json:"localAgentPort,omitempty" toml:"localAgentPort,omitempty" yaml:"localAgentPort,omitempty"`
	GlobalTag               string `description:"Sets a key:value tag on all spans." json:"globalTag,omitempty" toml:"globalTag,omitempty" yaml:"globalTag,omitempty" export:"true"`
	TraceIDHeaderName       string `description:"Sets the header name used to store the trace ID." json:"traceIDHeaderName,omitempty" toml:"traceIDHeaderName,omitempty" yaml:"traceIDHeaderName,omitempty" export:"true"`
	ParentIDHeaderName      string `description:"Sets the header name used to store the parent ID." json:"parentIDHeaderName,omitempty" toml:"parentIDHeaderName,omitempty" yaml:"parentIDHeaderName,omitempty" export:"true"`
	SpanIDHeaderName        string `description:"Sets the header name used to store the span ID." json:"spanIDHeaderName,omitempty" toml:"spanIDHeaderName,omitempty" yaml:"spanIDHeaderName,omitempty" export:"true"`
	BaggagePrefixHeaderName string `description:"Sets the header name prefix used to store baggage items in a map." json:"baggagePrefixHeaderName,omitempty" toml:"baggagePrefixHeaderName,omitempty" yaml:"baggagePrefixHeaderName,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.LocalAgentHost = "127.0.0.1"
	c.LocalAgentPort = 35000
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	tag := strings.SplitN(c.GlobalTag, ":", 2)

	value := ""
	if len(tag) == 2 {
		value = tag[1]
	}

	host := "localhost"
	port := 35000
	if len(c.LocalAgentHost) > 0 {
		host = c.LocalAgentHost
	}
	if c.LocalAgentPort > 0 {
		port = c.LocalAgentPort
	}

	tracer, closer := haystack.NewTracer(serviceName, haystack.NewAgentDispatcher(host, port, 3*time.Second, 1000),
		haystack.TracerOptionsFactory.Tag(tag[0], value),
		haystack.TracerOptionsFactory.Propagator(opentracing.HTTPHeaders,
			haystack.NewTextMapPropagator(haystack.PropagatorOpts{
				TraceIDKEYName:       c.TraceIDHeaderName,
				ParentSpanIDKEYName:  c.ParentIDHeaderName,
				SpanIDKEYName:        c.SpanIDHeaderName,
				BaggagePrefixKEYName: c.BaggagePrefixHeaderName,
			}, haystack.DefaultCodex{})),
		haystack.TracerOptionsFactory.Logger(&haystackLogger{logger: log.WithoutContext()}),
	)

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("haystack tracer configured")

	return tracer, closer, nil
}
