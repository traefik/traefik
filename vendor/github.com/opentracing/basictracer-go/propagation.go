package basictracer

import opentracing "github.com/opentracing/opentracing-go"

type accessorPropagator struct {
	tracer *tracerImpl
}

// DelegatingCarrier is a flexible carrier interface which can be implemented
// by types which have a means of storing the trace metadata and already know
// how to serialize themselves (for example, protocol buffers).
type DelegatingCarrier interface {
	SetState(traceID, spanID uint64, sampled bool)
	State() (traceID, spanID uint64, sampled bool)
	SetBaggageItem(key, value string)
	GetBaggage(func(key, value string))
}

func (p *accessorPropagator) Inject(
	spanContext opentracing.SpanContext,
	carrier interface{},
) error {
	dc, ok := carrier.(DelegatingCarrier)
	if !ok || dc == nil {
		return opentracing.ErrInvalidCarrier
	}
	sc, ok := spanContext.(SpanContext)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	dc.SetState(sc.TraceID, sc.SpanID, sc.Sampled)
	for k, v := range sc.Baggage {
		dc.SetBaggageItem(k, v)
	}
	return nil
}

func (p *accessorPropagator) Extract(
	carrier interface{},
) (opentracing.SpanContext, error) {
	dc, ok := carrier.(DelegatingCarrier)
	if !ok || dc == nil {
		return nil, opentracing.ErrInvalidCarrier
	}

	traceID, spanID, sampled := dc.State()
	sc := SpanContext{
		TraceID: traceID,
		SpanID:  spanID,
		Sampled: sampled,
		Baggage: nil,
	}
	dc.GetBaggage(func(k, v string) {
		if sc.Baggage == nil {
			sc.Baggage = map[string]string{}
		}
		sc.Baggage[k] = v
	})

	return sc, nil
}
