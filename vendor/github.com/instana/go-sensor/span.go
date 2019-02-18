package instana

import (
	"fmt"
	"os"
	"sync"
	"time"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

type spanS struct {
	tracer *tracerS
	sync.Mutex

	context      SpanContext
	ParentSpanID int64
	Operation    string
	Start        time.Time
	Duration     time.Duration
	Tags         ot.Tags
	Logs         []ot.LogRecord
	Error        bool
	Ec           int
}

func (r *spanS) BaggageItem(key string) string {
	r.Lock()
	defer r.Unlock()

	return r.context.Baggage[key]
}

func (r *spanS) SetBaggageItem(key, val string) ot.Span {
	if r.trim() {
		return r
	}

	r.Lock()
	defer r.Unlock()
	r.context = r.context.WithBaggageItem(key, val)

	return r
}

func (r *spanS) Context() ot.SpanContext {
	return r.context
}

func (r *spanS) Finish() {
	r.FinishWithOptions(ot.FinishOptions{})
}

func (r *spanS) FinishWithOptions(opts ot.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}

	duration := finishTime.Sub(r.Start)
	r.Lock()
	defer r.Unlock()
	for _, lr := range opts.LogRecords {
		r.appendLog(lr)
	}

	for _, ld := range opts.BulkLogData {
		r.appendLog(ld.ToLogRecord())
	}

	r.Duration = duration
	r.tracer.options.Recorder.RecordSpan(r)
}

func (r *spanS) appendLog(lr ot.LogRecord) {
	maxLogs := r.tracer.options.MaxLogsPerSpan
	if maxLogs == 0 || len(r.Logs) < maxLogs {
		r.Logs = append(r.Logs, lr)
	}
}

func (r *spanS) Log(ld ot.LogData) {
	r.Lock()
	defer r.Unlock()
	if r.trim() || r.tracer.options.DropAllLogs {
		return
	}

	if ld.Timestamp.IsZero() {
		ld.Timestamp = time.Now()
	}

	r.appendLog(ld.ToLogRecord())
}

func (r *spanS) trim() bool {
	return !r.context.Sampled && r.tracer.options.TrimUnsampledSpans
}

func (r *spanS) LogEvent(event string) {
	r.Log(ot.LogData{
		Event: event})
}

func (r *spanS) LogEventWithPayload(event string, payload interface{}) {
	r.Log(ot.LogData{
		Event:   event,
		Payload: payload})
}

func (r *spanS) LogFields(fields ...otlog.Field) {

	for _, v := range fields {
		// If this tag indicates an error, increase the error count
		if v.Key() == "error" {
			r.Error = true
			r.Ec++
		}
	}

	lr := ot.LogRecord{
		Fields: fields,
	}

	r.Lock()
	defer r.Unlock()
	if r.trim() || r.tracer.options.DropAllLogs {
		return
	}

	if lr.Timestamp.IsZero() {
		lr.Timestamp = time.Now()
	}

	r.appendLog(lr)
}

func (r *spanS) LogKV(keyValues ...interface{}) {
	fields, err := otlog.InterleavedKVToFields(keyValues...)
	if err != nil {
		r.LogFields(otlog.Error(err), otlog.String("function", "LogKV"))

		return
	}

	r.LogFields(fields...)
}

func (r *spanS) SetOperationName(operationName string) ot.Span {
	r.Lock()
	defer r.Unlock()
	r.Operation = operationName

	return r
}

func (r *spanS) SetTag(key string, value interface{}) ot.Span {
	r.Lock()
	defer r.Unlock()
	if r.trim() {
		return r
	}

	if r.Tags == nil {
		r.Tags = ot.Tags{}
	}

	// If this tag indicates an error, increase the error count
	if key == "error" {
		r.Error = true
		r.Ec++
	}

	r.Tags[key] = value

	return r
}

func (r *spanS) Tracer() ot.Tracer {
	return r.tracer
}

func (r *spanS) getTag(tag string) interface{} {
	var x, ok = r.Tags[tag]
	if !ok {
		x = ""
	}
	return x
}

func (r *spanS) getIntTag(tag string) int {
	d := r.Tags[tag]
	if d == nil {
		return -1
	}

	x, ok := d.(int)
	if !ok {
		return -1
	}

	return x
}

func (r *spanS) getStringTag(tag string) string {
	d := r.Tags[tag]
	if d == nil {
		return ""
	}
	return fmt.Sprint(d)
}

func (r *spanS) getHostName() string {
	hostTag := r.getStringTag(string(ext.PeerHostname))
	if hostTag != "" {
		return hostTag
	}

	h, err := os.Hostname()
	if err != nil {
		h = "localhost"
	}
	return h
}

func (r *spanS) getSpanKind() string {
	kind := r.getStringTag(string(ext.SpanKind))

	switch kind {
	case string(ext.SpanKindRPCServerEnum), "consumer", "entry":
		return "entry"
	case string(ext.SpanKindRPCClientEnum), "producer", "exit":
		return "exit"
	}
	return ""
}

func (r *spanS) collectLogs() map[uint64]map[string]interface{} {
	logs := make(map[uint64]map[string]interface{})
	for _, l := range r.Logs {
		if _, ok := logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)]; !ok {
			logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)] = make(map[string]interface{})
		}

		for _, f := range l.Fields {
			logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)][f.Key()] = f.Value()
		}
	}

	return logs
}
