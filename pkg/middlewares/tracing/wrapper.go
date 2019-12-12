package tracing

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

// Tracable embeds tracing information.
type Tracable interface {
	GetTracingInformation() (name string, spanKind ext.SpanKindEnum)
}

// Wrap adds tracability to an alice.Constructor.
func Wrap(ctx context.Context, constructor alice.Constructor) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if constructor == nil {
			return nil, nil
		}
		// we need to start tracing when middleware (provided by `constructor`) starts
		// and finish when middleware ends, but before calling `next`
		// so we need to wrap `next` into TraceFinisher
		// and call TraceFinisher when middleware processed
		tracingFinisher := NewTraceFinisher(next)
		handler, err := constructor(tracingFinisher)
		if err != nil {
			return nil, err
		}

		if tracableHandler, ok := handler.(Tracable); ok {
			name, spanKind := tracableHandler.GetTracingInformation()
			log.FromContext(ctx).WithField(log.MiddlewareName, name).Debug("Adding tracing to middleware")
			return NewTraceStarted(handler, name, spanKind, tracingFinisher), nil
		}
		return handler, nil
	}
}

// NewTraceStarted returns a *TracingStarted struct
func NewTraceStarted(middleware http.Handler, name string, spanKind ext.SpanKindEnum, finisher *TraceFinisher) *TraceStarter {
	return &TraceStarter{
		middleware: middleware,
		name:       name,
		spanKind:   spanKind,
		finisher:   finisher,
	}
}

// TraceStarter is used to start tracing for provided middleware and setup TraceFinisher to finish tracing
type TraceStarter struct {
	middleware http.Handler
	name       string
	spanKind   ext.SpanKindEnum
	finisher   *TraceFinisher
	finishFn   func()
	finished   bool
}

func (ts *TraceStarter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	_, err := tracing.FromContext(req.Context())
	if err != nil {
		if ts.middleware != nil {
			ts.middleware.ServeHTTP(rw, req)
		}
		return
	}

	_, req, ts.finishFn = tracing.StartSpan(req, ts.name, ts.spanKind)
	defer ts.finish()

	if ts.finisher != nil {
		ts.finisher.finish = ts.finish
	}

	if ts.middleware != nil {
		ts.middleware.ServeHTTP(rw, req)
	}
}

func (ts *TraceStarter) finish() {
	if ts.finished {
		return
	}
	ts.finishFn()
	ts.finished = true
}

// NewTraceFinisher returns a *TraceFinisher struct
func NewTraceFinisher(next http.Handler) *TraceFinisher {
	return &TraceFinisher{
		next: next,
	}
}

// TraceFinisher is used to finish tracing, started by TraceStarter
type TraceFinisher struct {
	next   http.Handler
	finish func()
}

func (tf *TraceFinisher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if tf.finish != nil {
		tf.finish()
	}

	if tf.next != nil {
		tf.next.ServeHTTP(rw, req)
	}
}
