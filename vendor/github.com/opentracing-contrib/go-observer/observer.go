// This project is licensed under the Apache License 2.0, see LICENSE.

package otobserver

import opentracing "github.com/opentracing/opentracing-go"

// Observer can be registered with a Tracer to recieve notifications
// about new Spans. Tracers are not required to support the Observer API.
// The actual registration depends on the implementation, which might look
// like the below e.g :
// observer := myobserver.NewObserver()
// tracer := client.NewTracer(..., client.WithObserver(observer))
//
type Observer interface {
	// Create and return a span observer. Called when a span starts.
	// If the Observer is not interested in the given span, it must return (nil, false).
	// E.g :
	//     func StartSpan(opName string, opts ...opentracing.StartSpanOption) {
	//         var sp opentracing.Span
	//         sso := opentracing.StartSpanOptions{}
	//         spanObserver, ok := observer.OnStartSpan(span, opName, sso);
	//         if ok {
	//             // we have a valid SpanObserver
	//         }
	//         ...
	//     }
	OnStartSpan(sp opentracing.Span, operationName string, options opentracing.StartSpanOptions) (SpanObserver, bool)
}

// SpanObserver is created by the Observer and receives notifications about
// other Span events.
type SpanObserver interface {
	// Callback called from opentracing.Span.SetOperationName()
	OnSetOperationName(operationName string)
	// Callback called from opentracing.Span.SetTag()
	OnSetTag(key string, value interface{})
	// Callback called from opentracing.Span.Finish()
	OnFinish(options opentracing.FinishOptions)
}
