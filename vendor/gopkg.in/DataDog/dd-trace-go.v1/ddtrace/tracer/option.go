package tracer

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/globalconfig"
)

// config holds the tracer configuration.
type config struct {
	// debug, when true, writes details to logs.
	debug bool

	// serviceName specifies the name of this application.
	serviceName string

	// sampler specifies the sampler that will be used for sampling traces.
	sampler Sampler

	// agentAddr specifies the hostname and  of the agent where the traces
	// are sent to.
	agentAddr string

	// globalTags holds a set of tags that will be automatically applied to
	// all spans.
	globalTags map[string]interface{}

	// transport specifies the Transport interface which will be used to send data to the agent.
	transport transport

	// propagator propagates span context cross-process
	propagator Propagator

	// httpRoundTripper defines the http.RoundTripper used by the agent transport.
	httpRoundTripper http.RoundTripper
}

// StartOption represents a function that can be provided as a parameter to Start.
type StartOption func(*config)

// defaults sets the default values for a config.
func defaults(c *config) {
	c.serviceName = filepath.Base(os.Args[0])
	c.sampler = NewAllSampler()
	c.agentAddr = defaultAddress
}

// WithPrioritySampling is deprecated, and priority sampling is enabled by default.
// When using distributed tracing, the priority sampling value is propagated in order to
// get all the parts of a distributed trace sampled.
// To learn more about priority sampling, please visit:
// https://docs.datadoghq.com/tracing/getting_further/trace_sampling_and_storage/#priority-sampling-for-distributed-tracing
func WithPrioritySampling() StartOption {
	return func(c *config) {
		// This is now enabled by default.
	}
}

// WithDebugMode enables debug mode on the tracer, resulting in more verbose logging.
func WithDebugMode(enabled bool) StartOption {
	return func(c *config) {
		c.debug = enabled
	}
}

// WithPropagator sets an alternative propagator to be used by the tracer.
func WithPropagator(p Propagator) StartOption {
	return func(c *config) {
		c.propagator = p
	}
}

// WithServiceName sets the default service name to be used with the tracer.
func WithServiceName(name string) StartOption {
	return func(c *config) {
		c.serviceName = name
	}
}

// WithAgentAddr sets the address where the agent is located. The default is
// localhost:8126. It should contain both host and port.
func WithAgentAddr(addr string) StartOption {
	return func(c *config) {
		c.agentAddr = addr
	}
}

// WithGlobalTag sets a key/value pair which will be set as a tag on all spans
// created by tracer. This option may be used multiple times.
func WithGlobalTag(k string, v interface{}) StartOption {
	return func(c *config) {
		if c.globalTags == nil {
			c.globalTags = make(map[string]interface{})
		}
		c.globalTags[k] = v
	}
}

// WithSampler sets the given sampler to be used with the tracer. By default
// an all-permissive sampler is used.
func WithSampler(s Sampler) StartOption {
	return func(c *config) {
		c.sampler = s
	}
}

// WithHTTPRoundTripper allows customizing the underlying HTTP transport for
// emitting spans. This is useful for advanced customization such as emitting
// spans to a unix domain socket. The default should be used in most cases.
func WithHTTPRoundTripper(r http.RoundTripper) StartOption {
	return func(c *config) {
		c.httpRoundTripper = r
	}
}

// WithAnalytics allows specifying whether Trace Search & Analytics should be enabled
// for integrations.
func WithAnalytics(on bool) StartOption {
	if on {
		return WithAnalyticsRate(1.0)
	}
	return WithAnalyticsRate(0.0)
}

// WithAnalyticsRate sets the global sampling rate for sampling APM events.
func WithAnalyticsRate(rate float64) StartOption {
	return func(_ *config) {
		globalconfig.SetAnalyticsRate(rate)
	}
}

// StartSpanOption is a configuration option for StartSpan. It is aliased in order
// to help godoc group all the functions returning it together. It is considered
// more correct to refer to it as the type as the origin, ddtrace.StartSpanOption.
type StartSpanOption = ddtrace.StartSpanOption

// Tag sets the given key/value pair as a tag on the started Span.
func Tag(k string, v interface{}) StartSpanOption {
	return func(cfg *ddtrace.StartSpanConfig) {
		if cfg.Tags == nil {
			cfg.Tags = map[string]interface{}{}
		}
		cfg.Tags[k] = v
	}
}

// ServiceName sets the given service name on the started span. For example "http.server".
func ServiceName(name string) StartSpanOption {
	return Tag(ext.ServiceName, name)
}

// ResourceName sets the given resource name on the started span. A resource could
// be an SQL query, a URL, an RPC method or something else.
func ResourceName(name string) StartSpanOption {
	return Tag(ext.ResourceName, name)
}

// SpanType sets the given span type on the started span. Some examples in the case of
// the Datadog APM product could be "web", "db" or "cache".
func SpanType(name string) StartSpanOption {
	return Tag(ext.SpanType, name)
}

// WithSpanID sets the SpanID on the started span, instead of using a random number.
// If there is no parent Span (eg from ChildOf), then the TraceID will also be set to the
// value given here.
func WithSpanID(id uint64) StartSpanOption {
	return func(cfg *ddtrace.StartSpanConfig) {
		cfg.SpanID = id
	}
}

// ChildOf tells StartSpan to use the given span context as a parent for the
// created span.
func ChildOf(ctx ddtrace.SpanContext) StartSpanOption {
	return func(cfg *ddtrace.StartSpanConfig) {
		cfg.Parent = ctx
	}
}

// StartTime sets a custom time as the start time for the created span. By
// default a span is started using the creation time.
func StartTime(t time.Time) StartSpanOption {
	return func(cfg *ddtrace.StartSpanConfig) {
		cfg.StartTime = t
	}
}

// FinishOption is a configuration option for FinishSpan. It is aliased in order
// to help godoc group all the functions returning it together. It is considered
// more correct to refer to it as the type as the origin, ddtrace.FinishOption.
type FinishOption = ddtrace.FinishOption

// FinishTime sets the given time as the finishing time for the span. By default,
// the current time is used.
func FinishTime(t time.Time) FinishOption {
	return func(cfg *ddtrace.FinishConfig) {
		cfg.FinishTime = t
	}
}

// WithError marks the span as having had an error. It uses the information from
// err to set tags such as the error message, error type and stack trace. It has
// no effect if the error is nil.
func WithError(err error) FinishOption {
	return func(cfg *ddtrace.FinishConfig) {
		cfg.Error = err
	}
}

// NoDebugStack prevents any error presented using the WithError finishing option
// from generating a stack trace. This is useful in situations where errors are frequent
// and performance is critical.
func NoDebugStack() FinishOption {
	return func(cfg *ddtrace.FinishConfig) {
		cfg.NoDebugStack = true
	}
}
