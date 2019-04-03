package wire

import (
	"github.com/openzipkin-contrib/zipkin-go-opentracing/flag"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/types"
)

// ProtobufCarrier is a DelegatingCarrier that uses protocol buffers as the
// the underlying datastructure. The reason for implementing DelagatingCarrier
// is to allow for end users to serialize the underlying protocol buffers using
// jsonpb or any other serialization forms they want.
type ProtobufCarrier TracerState

// SetState set's the tracer state.
func (p *ProtobufCarrier) SetState(traceID types.TraceID, spanID uint64, parentSpanID *uint64, sampled bool, flags flag.Flags) {
	p.TraceId = traceID.Low
	p.TraceIdHigh = traceID.High
	p.SpanId = spanID
	if parentSpanID == nil {
		flags |= flag.IsRoot
		p.ParentSpanId = 0
	} else {
		flags &^= flag.IsRoot
		p.ParentSpanId = *parentSpanID
	}
	flags |= flag.SamplingSet
	if sampled {
		flags |= flag.Sampled
		p.Sampled = sampled
	} else {
		flags &^= flag.Sampled
	}
	p.Flags = uint64(flags)
}

// State returns the tracer state.
func (p *ProtobufCarrier) State() (traceID types.TraceID, spanID uint64, parentSpanID *uint64, sampled bool, flags flag.Flags) {
	traceID.Low = p.TraceId
	traceID.High = p.TraceIdHigh
	spanID = p.SpanId
	sampled = p.Sampled
	flags = flag.Flags(p.Flags)
	if flags&flag.IsRoot == 0 {
		parentSpanID = &p.ParentSpanId
	}
	return traceID, spanID, parentSpanID, sampled, flags
}

// SetBaggageItem sets a baggage item.
func (p *ProtobufCarrier) SetBaggageItem(key, value string) {
	if p.BaggageItems == nil {
		p.BaggageItems = map[string]string{key: value}
		return
	}

	p.BaggageItems[key] = value
}

// GetBaggage iterates over each baggage item and executes the callback with
// the key:value pair.
func (p *ProtobufCarrier) GetBaggage(f func(k, v string)) {
	for k, v := range p.BaggageItems {
		f(k, v)
	}
}
