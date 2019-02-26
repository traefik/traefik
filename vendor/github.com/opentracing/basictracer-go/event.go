package basictracer

import "github.com/opentracing/opentracing-go"

// A SpanEvent is emitted when a mutating command is called on a Span.
type SpanEvent interface{}

// EventCreate is emitted when a Span is created.
type EventCreate struct{ OperationName string }

// EventTag is received when SetTag is called.
type EventTag struct {
	Key   string
	Value interface{}
}

// EventBaggage is received when SetBaggageItem is called.
type EventBaggage struct {
	Key, Value string
}

// EventLogFields is received when LogFields or LogKV is called.
type EventLogFields opentracing.LogRecord

// EventLog is received when Log (or one of its derivatives) is called.
//
// DEPRECATED
type EventLog opentracing.LogData

// EventFinish is received when Finish is called.
type EventFinish RawSpan

func (s *spanImpl) onCreate(opName string) {
	if s.event != nil {
		s.event(EventCreate{OperationName: opName})
	}
}
func (s *spanImpl) onTag(key string, value interface{}) {
	if s.event != nil {
		s.event(EventTag{Key: key, Value: value})
	}
}
func (s *spanImpl) onLog(ld opentracing.LogData) {
	if s.event != nil {
		s.event(EventLog(ld))
	}
}
func (s *spanImpl) onLogFields(lr opentracing.LogRecord) {
	if s.event != nil {
		s.event(EventLogFields(lr))
	}
}
func (s *spanImpl) onBaggage(key, value string) {
	if s.event != nil {
		s.event(EventBaggage{Key: key, Value: value})
	}
}
func (s *spanImpl) onFinish(sp RawSpan) {
	if s.event != nil {
		s.event(EventFinish(sp))
	}
}
