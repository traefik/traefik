package haystack

import (
	"io"
	"strings"
	"time"

	"github.com/ExpediaDotCom/haystack-client-go"
	"github.com/containous/traefik/pkg/log"
	"github.com/opentracing/opentracing-go"
)

// Name sets the name of this tracer
const Name = "haystack"

// Config provides configuration settings for a haystack tracer
type Config struct {
	LocalAgentHost          string `description:"Set haystack-agent's host that the reporter will used." export:"false"`
	LocalAgentPort          int    `description:"Set haystack-agent's port that the reporter will used." export:"false"`
	GlobalTag               string `description:"Key:Value tag to be set on all the spans." export:"true"`
	TraceIDHeaderName       string `description:"Specifies the header name that will be used to store the trace ID." export:"true"`
	ParentIDHeaderName      string `description:"Specifies the header name that will be used to store the parent ID." export:"true"`
	SpanIDHeaderName        string `description:"Specifies the header name that will be used to store the span ID." export:"true"`
	BaggagePrefixHeaderName string `description:"specifies the header name prefix that will be used to store baggage items in a map." export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.LocalAgentHost = "LocalAgentHost"
	c.LocalAgentPort = 35000
}

// Setup sets up the tracer
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

	log.WithoutContext().Debug("DataDog tracer configured")

	return tracer, closer, nil
}
