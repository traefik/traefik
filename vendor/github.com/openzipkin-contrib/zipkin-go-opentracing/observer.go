package zipkintracer

import (
	otobserver "github.com/opentracing-contrib/go-observer"
	opentracing "github.com/opentracing/opentracing-go"
)

// observer is a dispatcher to other observers
type observer struct {
	observers []otobserver.Observer
}

// spanObserver is a dispatcher to other span observers
type spanObserver struct {
	observers []otobserver.SpanObserver
}

func (o observer) OnStartSpan(sp opentracing.Span, operationName string, options opentracing.StartSpanOptions) (otobserver.SpanObserver, bool) {
	var spanObservers []otobserver.SpanObserver
	for _, obs := range o.observers {
		spanObs, ok := obs.OnStartSpan(sp, operationName, options)
		if ok {
			if spanObservers == nil {
				spanObservers = make([]otobserver.SpanObserver, 0, len(o.observers))
			}
			spanObservers = append(spanObservers, spanObs)
		}
	}
	if len(spanObservers) == 0 {
		return nil, false
	}

	return spanObserver{observers: spanObservers}, true
}

func (o spanObserver) OnSetOperationName(operationName string) {
	for _, obs := range o.observers {
		obs.OnSetOperationName(operationName)
	}
}

func (o spanObserver) OnSetTag(key string, value interface{}) {
	for _, obs := range o.observers {
		obs.OnSetTag(key, value)
	}
}

func (o spanObserver) OnFinish(options opentracing.FinishOptions) {
	for _, obs := range o.observers {
		obs.OnFinish(options)
	}
}
