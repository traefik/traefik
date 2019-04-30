package haystack

import (
	"io"
	"strconv"
	"strings"
	"time"

	haystack "github.com/ExpediaDotCom/haystack-client-go"
	"github.com/containous/traefik/pkg/log"
	"github.com/opentracing/opentracing-go"
)

// Name sets the name of this tracer
const Name = "haystack"

// Config provides configuration settings for a datadog tracer
type Config struct {
	LocalAgentHostPort     string `description:"Set datadog-agent's host:port that the reporter will used. Defaults to localhost:35000" export:"false"`
	GlobalTag              string `description:"Key:Value tag to be set on all the spans." export:"true"`
	TraceIDHeaderName      string `description:"Specifies the header name that will be used to store the trace ID.." export:"true"`
	ParentIDHeaderName     string `description:"Specifies the header name that will be used to store the parent ID." export:"true"`
	SpanIDHeaderName       string `description:"Specifies the header name that will be used to store the span ID." export:"true"`
	BagagePrefixHeaderName string `description:"specifies the header name prefix that will be used to store baggage items in a map." export:"true"`
}

// Setup sets up the tracer
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	tag := strings.SplitN(c.GlobalTag, ":", 2)

	value := ""
	if len(tag) == 2 {
		value = tag[1]
	}

	hostAndPort := strings.SplitN(c.LocalAgentHostPort, ":", 2)
	host := "localhost"
	port := 35000
	if len(hostAndPort) > 0 {
		host = hostAndPort[0]
		if len(hostAndPort) == 2 {
			port, _ = strconv.Atoi(hostAndPort[1])
		}
	}

	tracer, closer := haystack.NewTracer(serviceName, haystack.NewAgentDispatcher(host, port, 3*time.Second, 1000),
		haystack.TracerOptionsFactory.Tag(tag[0], value),
		haystack.TracerOptionsFactory.Propagator(opentracing.HTTPHeaders,
			haystack.NewTextMapPropagator(haystack.PropagatorOpts{
				TraceIDKEYName:       c.TraceIDHeaderName,
				ParentSpanIDKEYName:  c.ParentIDHeaderName,
				SpanIDKEYName:        c.SpanIDHeaderName,
				BaggagePrefixKEYName: c.BagagePrefixHeaderName,
			}, haystack.DefaultCodex{})),
		haystack.TracerOptionsFactory.Logger(&loggerWrapper{logger: log.WithoutContext()}),
	)

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("DataDog tracer configured")

	return tracer, closer, nil
}

/*NullLogger does nothing*/
type loggerWrapper struct {
	logger log.Logger
}

/*Error prints the error message*/
func (l loggerWrapper) Error(format string, v ...interface{}) {
	l.logger.Errorf(format, v)
}

/*Info prints the info message*/
func (l loggerWrapper) Info(format string, v ...interface{}) {
	l.logger.Infof(format, v)
}

/*Debug prints the info message*/
func (l loggerWrapper) Debug(format string, v ...interface{}) {
	l.logger.Debug(format, v)
}
