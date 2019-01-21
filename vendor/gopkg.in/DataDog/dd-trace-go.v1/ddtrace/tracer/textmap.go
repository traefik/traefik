package tracer

import (
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
)

// HTTPHeadersCarrier wraps an http.Header as a TextMapWriter and TextMapReader, allowing
// it to be used using the provided Propagator implementation.
type HTTPHeadersCarrier http.Header

var _ TextMapWriter = (*HTTPHeadersCarrier)(nil)
var _ TextMapReader = (*HTTPHeadersCarrier)(nil)

// Set implements TextMapWriter.
func (c HTTPHeadersCarrier) Set(key, val string) {
	http.Header(c).Set(key, val)
}

// ForeachKey implements TextMapReader.
func (c HTTPHeadersCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range c {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// TextMapCarrier allows the use of a regular map[string]string as both TextMapWriter
// and TextMapReader, making it compatible with the provided Propagator.
type TextMapCarrier map[string]string

var _ TextMapWriter = (*TextMapCarrier)(nil)
var _ TextMapReader = (*TextMapCarrier)(nil)

// Set implements TextMapWriter.
func (c TextMapCarrier) Set(key, val string) {
	c[key] = val
}

// ForeachKey conforms to the TextMapReader interface.
func (c TextMapCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, v := range c {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

const (
	// DefaultBaggageHeaderPrefix specifies the prefix that will be used in
	// HTTP headers or text maps to prefix baggage keys.
	DefaultBaggageHeaderPrefix = "ot-baggage-"

	// DefaultTraceIDHeader specifies the key that will be used in HTTP headers
	// or text maps to store the trace ID.
	DefaultTraceIDHeader = "x-datadog-trace-id"

	// DefaultParentIDHeader specifies the key that will be used in HTTP headers
	// or text maps to store the parent ID.
	DefaultParentIDHeader = "x-datadog-parent-id"

	// DefaultPriorityHeader specifies the key that will be used in HTTP headers
	// or text maps to store the sampling priority value.
	DefaultPriorityHeader = "x-datadog-sampling-priority"
)

// PropagatorConfig defines the configuration for initializing a propagator.
type PropagatorConfig struct {
	// BaggagePrefix specifies the prefix that will be used to store baggage
	// items in a map. It defaults to DefaultBaggageHeaderPrefix.
	BaggagePrefix string

	// TraceHeader specifies the map key that will be used to store the trace ID.
	// It defaults to DefaultTraceIDHeader.
	TraceHeader string

	// ParentHeader specifies the map key that will be used to store the parent ID.
	// It defaults to DefaultParentIDHeader.
	ParentHeader string

	// PriorityHeader specifies the map key that will be used to store the sampling priority.
	// It deafults to DefaultPriorityHeader.
	PriorityHeader string
}

// NewPropagator returns a new propagator which uses TextMap to inject
// and extract values. It propagates trace and span IDs and baggage.
// To use the defaults, nil may be provided in place of the config.
func NewPropagator(cfg *PropagatorConfig) Propagator {
	if cfg == nil {
		cfg = new(PropagatorConfig)
	}
	if cfg.BaggagePrefix == "" {
		cfg.BaggagePrefix = DefaultBaggageHeaderPrefix
	}
	if cfg.TraceHeader == "" {
		cfg.TraceHeader = DefaultTraceIDHeader
	}
	if cfg.ParentHeader == "" {
		cfg.ParentHeader = DefaultParentIDHeader
	}
	if cfg.PriorityHeader == "" {
		cfg.PriorityHeader = DefaultPriorityHeader
	}
	return &propagator{cfg}
}

// propagator implements a propagator which uses TextMap internally.
// It propagates the trace and span IDs, as well as the baggage from the
// context.
type propagator struct{ cfg *PropagatorConfig }

// Inject defines the Propagator to propagate SpanContext data
// out of the current process. The implementation propagates the
// TraceID and the current active SpanID, as well as the Span baggage.
func (p *propagator) Inject(spanCtx ddtrace.SpanContext, carrier interface{}) error {
	switch v := carrier.(type) {
	case TextMapWriter:
		return p.injectTextMap(spanCtx, v)
	default:
		return ErrInvalidCarrier
	}
}

func (p *propagator) injectTextMap(spanCtx ddtrace.SpanContext, writer TextMapWriter) error {
	ctx, ok := spanCtx.(*spanContext)
	if !ok || ctx.traceID == 0 || ctx.spanID == 0 {
		return ErrInvalidSpanContext
	}
	// propagate the TraceID and the current active SpanID
	writer.Set(p.cfg.TraceHeader, strconv.FormatUint(ctx.traceID, 10))
	writer.Set(p.cfg.ParentHeader, strconv.FormatUint(ctx.spanID, 10))
	if ctx.hasSamplingPriority() {
		writer.Set(p.cfg.PriorityHeader, strconv.Itoa(ctx.samplingPriority()))
	}
	// propagate OpenTracing baggage
	for k, v := range ctx.baggage {
		writer.Set(p.cfg.BaggagePrefix+k, v)
	}
	return nil
}

// Extract implements Propagator.
func (p *propagator) Extract(carrier interface{}) (ddtrace.SpanContext, error) {
	switch v := carrier.(type) {
	case TextMapReader:
		return p.extractTextMap(v)
	default:
		return nil, ErrInvalidCarrier
	}
}

func (p *propagator) extractTextMap(reader TextMapReader) (ddtrace.SpanContext, error) {
	var ctx spanContext
	err := reader.ForeachKey(func(k, v string) error {
		var err error
		key := strings.ToLower(k)
		switch key {
		case p.cfg.TraceHeader:
			ctx.traceID, err = parseUint64(v)
			if err != nil {
				return ErrSpanContextCorrupted
			}
		case p.cfg.ParentHeader:
			ctx.spanID, err = parseUint64(v)
			if err != nil {
				return ErrSpanContextCorrupted
			}
		case p.cfg.PriorityHeader:
			priority, err := strconv.Atoi(v)
			if err != nil {
				return ErrSpanContextCorrupted
			}
			ctx.setSamplingPriority(priority)
		default:
			if strings.HasPrefix(key, p.cfg.BaggagePrefix) {
				ctx.setBaggageItem(strings.TrimPrefix(key, p.cfg.BaggagePrefix), v)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if ctx.traceID == 0 || ctx.spanID == 0 {
		return nil, ErrSpanContextNotFound
	}
	return &ctx, nil
}
