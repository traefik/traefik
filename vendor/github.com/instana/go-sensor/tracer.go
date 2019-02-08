package instana

import (
	"time"

	ot "github.com/opentracing/opentracing-go"
)

const (
	// MaxLogsPerSpan The maximum number of logs allowed on a span.
	MaxLogsPerSpan = 2
)

type tracerS struct {
	options        TracerOptions
	textPropagator *textMapPropagator
}

func (r *tracerS) Inject(sc ot.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case ot.TextMap, ot.HTTPHeaders:
		return r.textPropagator.inject(sc, carrier)
	}

	return ot.ErrUnsupportedFormat
}

func (r *tracerS) Extract(format interface{}, carrier interface{}) (ot.SpanContext, error) {
	switch format {
	case ot.TextMap, ot.HTTPHeaders:
		return r.textPropagator.extract(carrier)
	}

	return nil, ot.ErrUnsupportedFormat
}

func (r *tracerS) StartSpan(operationName string, opts ...ot.StartSpanOption) ot.Span {
	sso := ot.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}

	return r.StartSpanWithOptions(operationName, sso)
}

func (r *tracerS) StartSpanWithOptions(operationName string, opts ot.StartSpanOptions) ot.Span {
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	tags := opts.Tags
	span := &spanS{}
Loop:
	for _, ref := range opts.References {
		switch ref.Type {
		case ot.ChildOfRef, ot.FollowsFromRef:
			refCtx := ref.ReferencedContext.(SpanContext)
			span.context.TraceID = refCtx.TraceID
			span.context.SpanID = randomID()
			span.context.Sampled = refCtx.Sampled
			span.ParentSpanID = refCtx.SpanID
			if l := len(refCtx.Baggage); l > 0 {
				span.context.Baggage = make(map[string]string, l)
				for k, v := range refCtx.Baggage {
					span.context.Baggage[k] = v
				}
			}

			break Loop
		}
	}

	if span.context.TraceID == 0 {
		span.context.SpanID = randomID()
		span.context.TraceID = span.context.SpanID
		span.context.Sampled = r.options.ShouldSample(span.context.TraceID)
	}

	return r.startSpanInternal(span, operationName, startTime, tags)
}

func (r *tracerS) startSpanInternal(span *spanS, operationName string, startTime time.Time, tags ot.Tags) ot.Span {
	span.tracer = r
	span.Operation = operationName
	span.Start = startTime
	span.Duration = -1
	span.Tags = tags

	return span
}

func shouldSample(traceID int64) bool {
	return false
}

// NewTracer Get a new Tracer with the default options applied.
func NewTracer() ot.Tracer {
	return NewTracerWithOptions(&Options{})
}

// NewTracerWithOptions Get a new Tracer with the specified options.
func NewTracerWithOptions(options *Options) ot.Tracer {
	InitSensor(options)

	return NewTracerWithEverything(options, NewRecorder())
}

// NewTracerWithEverything Get a new Tracer with the works.
func NewTracerWithEverything(options *Options, recorder SpanRecorder) ot.Tracer {
	InitSensor(options)
	ret := &tracerS{options: TracerOptions{
		Recorder:       recorder,
		ShouldSample:   shouldSample,
		MaxLogsPerSpan: MaxLogsPerSpan}}
	ret.textPropagator = &textMapPropagator{ret}

	return ret
}
