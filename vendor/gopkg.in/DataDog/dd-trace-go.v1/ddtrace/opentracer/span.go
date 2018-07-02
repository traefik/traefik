package opentracer

import (
	"fmt"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

var _ opentracing.Span = (*span)(nil)

// span implements opentracing.Span on top of ddtrace.Span.
type span struct {
	ddtrace.Span
	*opentracer
}

func (s *span) Context() opentracing.SpanContext                      { return s.Span.Context() }
func (s *span) Finish()                                               { s.Span.Finish() }
func (s *span) Tracer() opentracing.Tracer                            { return s.opentracer }
func (s *span) LogEvent(event string)                                 { /* deprecated */ }
func (s *span) LogEventWithPayload(event string, payload interface{}) { /* deprecated */ }
func (s *span) Log(data opentracing.LogData)                          { /* deprecated */ }

func (s *span) FinishWithOptions(opts opentracing.FinishOptions) {
	for _, lr := range opts.LogRecords {
		if len(lr.Fields) > 0 {
			s.LogFields(lr.Fields...)
		}
	}
	s.Span.Finish(tracer.FinishTime(opts.FinishTime))
}

func (s *span) LogFields(fields ...log.Field) {
	// catch standard opentracing keys and adjust to internal ones as per spec:
	// https://github.com/opentracing/specification/blob/master/semantic_conventions.md#log-fields-table
	for _, f := range fields {
		switch f.Key() {
		case "event":
			if v, ok := f.Value().(string); ok && v == "error" {
				s.SetTag("error", true)
			}
		case "error", "error.object":
			if err, ok := f.Value().(error); ok {
				s.SetTag("error", err)
			}
		case "message":
			s.SetTag(ext.ErrorMsg, fmt.Sprint(f.Value()))
		case "stack":
			s.SetTag(ext.ErrorStack, fmt.Sprint(f.Value()))
		default:
			// not implemented
		}
	}
}

func (s *span) LogKV(keyVals ...interface{}) {
	fields, err := log.InterleavedKVToFields(keyVals...)
	if err != nil {
		// TODO(gbbr): create a log package
		return
	}
	s.LogFields(fields...)
}

func (s *span) SetBaggageItem(key, val string) opentracing.Span {
	s.Span.SetBaggageItem(key, val)
	return s
}

func (s *span) SetOperationName(operationName string) opentracing.Span {
	s.Span.SetOperationName(operationName)
	return s
}

func (s *span) SetTag(key string, value interface{}) opentracing.Span {
	s.Span.SetTag(key, value)
	return s
}
