package zipkintracer

import (
	"github.com/openzipkin-contrib/zipkin-go-opentracing/flag"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/types"
)

// SpanContext holds the basic Span metadata.
type SpanContext struct {
	// A probabilistically unique identifier for a [multi-span] trace.
	TraceID types.TraceID

	// A probabilistically unique identifier for a span.
	SpanID uint64

	// Whether the trace is sampled.
	Sampled bool

	// The span's associated baggage.
	Baggage map[string]string // initialized on first use

	// The SpanID of this Context's parent, or nil if there is no parent.
	ParentSpanID *uint64

	// Flags provides the ability to create and communicate feature flags.
	Flags flag.Flags

	// Whether the span is owned by the current process
	Owner bool
}

// ForeachBaggageItem belongs to the opentracing.SpanContext interface
func (c SpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range c.Baggage {
		if !handler(k, v) {
			break
		}
	}
}

// WithBaggageItem returns an entirely new basictracer SpanContext with the
// given key:value baggage pair set.
func (c SpanContext) WithBaggageItem(key, val string) SpanContext {
	var newBaggage map[string]string
	if c.Baggage == nil {
		newBaggage = map[string]string{key: val}
	} else {
		newBaggage = make(map[string]string, len(c.Baggage)+1)
		for k, v := range c.Baggage {
			newBaggage[k] = v
		}
		newBaggage[key] = val
	}
	var parentSpanID *uint64
	if c.ParentSpanID != nil {
		parentSpanID = new(uint64)
		*parentSpanID = *c.ParentSpanID
	}
	// Use positional parameters so the compiler will help catch new fields.
	return SpanContext{c.TraceID, c.SpanID, c.Sampled, newBaggage, parentSpanID, c.Flags, c.Owner}
}
