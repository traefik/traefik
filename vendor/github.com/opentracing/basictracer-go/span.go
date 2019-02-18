package basictracer

import (
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// Span provides access to the essential details of the span, for use
// by basictracer consumers.  These methods may only be called prior
// to (*opentracing.Span).Finish().
type Span interface {
	opentracing.Span

	// Operation names the work done by this span instance
	Operation() string

	// Start indicates when the span began
	Start() time.Time
}

// Implements the `Span` interface. Created via tracerImpl (see
// `basictracer.New()`).
type spanImpl struct {
	tracer     *tracerImpl
	event      func(SpanEvent)
	sync.Mutex // protects the fields below
	raw        RawSpan
	// The number of logs dropped because of MaxLogsPerSpan.
	numDroppedLogs int
}

var spanPool = &sync.Pool{New: func() interface{} {
	return &spanImpl{}
}}

func (s *spanImpl) reset() {
	s.tracer, s.event = nil, nil
	// Note: Would like to do the following, but then the consumer of RawSpan
	// (the recorder) needs to make sure that they're not holding on to the
	// baggage or logs when they return (i.e. they need to copy if they care):
	//
	//     logs, baggage := s.raw.Logs[:0], s.raw.Baggage
	//     for k := range baggage {
	//         delete(baggage, k)
	//     }
	//     s.raw.Logs, s.raw.Baggage = logs, baggage
	//
	// That's likely too much to ask for. But there is some magic we should
	// be able to do with `runtime.SetFinalizer` to reclaim that memory into
	// a buffer pool when GC considers them unreachable, which should ease
	// some of the load. Hard to say how quickly that would be in practice
	// though.
	s.raw = RawSpan{
		Context: SpanContext{},
	}
}

func (s *spanImpl) SetOperationName(operationName string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.raw.Operation = operationName
	return s
}

func (s *spanImpl) trim() bool {
	return !s.raw.Context.Sampled && s.tracer.options.TrimUnsampledSpans
}

func (s *spanImpl) SetTag(key string, value interface{}) opentracing.Span {
	defer s.onTag(key, value)
	s.Lock()
	defer s.Unlock()
	if key == string(ext.SamplingPriority) {
		if v, ok := value.(uint16); ok {
			s.raw.Context.Sampled = v != 0
			return s
		}
	}
	if s.trim() {
		return s
	}

	if s.raw.Tags == nil {
		s.raw.Tags = opentracing.Tags{}
	}
	s.raw.Tags[key] = value
	return s
}

func (s *spanImpl) LogKV(keyValues ...interface{}) {
	fields, err := log.InterleavedKVToFields(keyValues...)
	if err != nil {
		s.LogFields(log.Error(err), log.String("function", "LogKV"))
		return
	}
	s.LogFields(fields...)
}

func (s *spanImpl) appendLog(lr opentracing.LogRecord) {
	maxLogs := s.tracer.options.MaxLogsPerSpan
	if maxLogs == 0 || len(s.raw.Logs) < maxLogs {
		s.raw.Logs = append(s.raw.Logs, lr)
		return
	}

	// We have too many logs. We don't touch the first numOld logs; we treat the
	// rest as a circular buffer and overwrite the oldest log among those.
	numOld := (maxLogs - 1) / 2
	numNew := maxLogs - numOld
	s.raw.Logs[numOld+s.numDroppedLogs%numNew] = lr
	s.numDroppedLogs++
}

func (s *spanImpl) LogFields(fields ...log.Field) {
	lr := opentracing.LogRecord{
		Fields: fields,
	}
	defer s.onLogFields(lr)
	s.Lock()
	defer s.Unlock()
	if s.trim() || s.tracer.options.DropAllLogs {
		return
	}
	if lr.Timestamp.IsZero() {
		lr.Timestamp = time.Now()
	}
	s.appendLog(lr)
}

func (s *spanImpl) LogEvent(event string) {
	s.Log(opentracing.LogData{
		Event: event,
	})
}

func (s *spanImpl) LogEventWithPayload(event string, payload interface{}) {
	s.Log(opentracing.LogData{
		Event:   event,
		Payload: payload,
	})
}

func (s *spanImpl) Log(ld opentracing.LogData) {
	defer s.onLog(ld)
	s.Lock()
	defer s.Unlock()
	if s.trim() || s.tracer.options.DropAllLogs {
		return
	}

	if ld.Timestamp.IsZero() {
		ld.Timestamp = time.Now()
	}

	s.appendLog(ld.ToLogRecord())
}

func (s *spanImpl) Finish() {
	s.FinishWithOptions(opentracing.FinishOptions{})
}

// rotateLogBuffer rotates the records in the buffer: records 0 to pos-1 move at
// the end (i.e. pos circular left shifts).
func rotateLogBuffer(buf []opentracing.LogRecord, pos int) {
	// This algorithm is described in:
	//    http://www.cplusplus.com/reference/algorithm/rotate
	for first, middle, next := 0, pos, pos; first != middle; {
		buf[first], buf[next] = buf[next], buf[first]
		first++
		next++
		if next == len(buf) {
			next = middle
		} else if first == middle {
			middle = next
		}
	}
}

func (s *spanImpl) FinishWithOptions(opts opentracing.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}
	duration := finishTime.Sub(s.raw.Start)

	s.Lock()
	defer s.Unlock()

	for _, lr := range opts.LogRecords {
		s.appendLog(lr)
	}
	for _, ld := range opts.BulkLogData {
		s.appendLog(ld.ToLogRecord())
	}

	if s.numDroppedLogs > 0 {
		// We dropped some log events, which means that we used part of Logs as a
		// circular buffer (see appendLog). De-circularize it.
		numOld := (len(s.raw.Logs) - 1) / 2
		numNew := len(s.raw.Logs) - numOld
		rotateLogBuffer(s.raw.Logs[numOld:], s.numDroppedLogs%numNew)

		// Replace the log in the middle (the oldest "new" log) with information
		// about the dropped logs. This means that we are effectively dropping one
		// more "new" log.
		numDropped := s.numDroppedLogs + 1
		s.raw.Logs[numOld] = opentracing.LogRecord{
			// Keep the timestamp of the last dropped event.
			Timestamp: s.raw.Logs[numOld].Timestamp,
			Fields: []log.Field{
				log.String("event", "dropped Span logs"),
				log.Int("dropped_log_count", numDropped),
				log.String("component", "basictracer"),
			},
		}
	}

	s.raw.Duration = duration

	s.onFinish(s.raw)
	s.tracer.options.Recorder.RecordSpan(s.raw)

	// Last chance to get options before the span is possibly reset.
	poolEnabled := s.tracer.options.EnableSpanPool
	if s.tracer.options.DebugAssertUseAfterFinish {
		// This makes it much more likely to catch a panic on any subsequent
		// operation since s.tracer is accessed on every call to `Lock`.
		// We don't call `reset()` here to preserve the logs in the Span
		// which are printed when the assertion triggers.
		s.tracer = nil
	}

	if poolEnabled {
		spanPool.Put(s)
	}
}

func (s *spanImpl) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *spanImpl) Context() opentracing.SpanContext {
	return s.raw.Context
}

func (s *spanImpl) SetBaggageItem(key, val string) opentracing.Span {
	s.onBaggage(key, val)
	if s.trim() {
		return s
	}

	s.Lock()
	defer s.Unlock()
	s.raw.Context = s.raw.Context.WithBaggageItem(key, val)
	return s
}

func (s *spanImpl) BaggageItem(key string) string {
	s.Lock()
	defer s.Unlock()
	return s.raw.Context.Baggage[key]
}

func (s *spanImpl) Operation() string {
	return s.raw.Operation
}

func (s *spanImpl) Start() time.Time {
	return s.raw.Start
}
