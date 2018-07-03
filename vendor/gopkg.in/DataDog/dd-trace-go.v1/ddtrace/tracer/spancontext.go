package tracer

import (
	"sync"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/internal"
)

var _ ddtrace.SpanContext = (*spanContext)(nil)

// SpanContext represents a span state that can propagate to descendant spans
// and across process boundaries. It contains all the information needed to
// spawn a direct descendant of the span that it belongs to. It can be used
// to create distributed tracing by propagating it using the provided interfaces.
type spanContext struct {
	// the below group should propagate only locally

	trace   *trace // reference to the trace that this span belongs too
	span    *span  // reference to the span that hosts this context
	sampled bool   // whether this span will be sampled or not

	// the below group should propagate cross-process

	traceID uint64
	spanID  uint64

	mu          sync.RWMutex // guards below fields
	baggage     map[string]string
	priority    int
	hasPriority bool
}

// newSpanContext creates a new SpanContext to serve as context for the given
// span. If the provided parent is not nil, the context will inherit the trace,
// baggage and other values from it. This method also pushes the span into the
// new context's trace and as a result, it should not be called multiple times
// for the same span.
func newSpanContext(span *span, parent *spanContext) *spanContext {
	context := &spanContext{
		traceID: span.TraceID,
		spanID:  span.SpanID,
		sampled: true,
		span:    span,
	}
	if v, ok := span.Metrics[samplingPriorityKey]; ok {
		context.hasPriority = true
		context.priority = int(v)
	}
	if parent != nil {
		context.trace = parent.trace
		context.sampled = parent.sampled
		context.hasPriority = parent.hasSamplingPriority()
		context.priority = parent.samplingPriority()
		parent.ForeachBaggageItem(func(k, v string) bool {
			context.setBaggageItem(k, v)
			return true
		})
	}
	if context.trace == nil {
		context.trace = newTrace()
	}
	// put span in context's trace
	context.trace.push(span)
	return context
}

// SpanID implements ddtrace.SpanContext.
func (c *spanContext) SpanID() uint64 { return c.spanID }

// TraceID implements ddtrace.SpanContext.
func (c *spanContext) TraceID() uint64 { return c.traceID }

// ForeachBaggageItem implements ddtrace.SpanContext.
func (c *spanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for k, v := range c.baggage {
		if !handler(k, v) {
			break
		}
	}
}

func (c *spanContext) setSamplingPriority(p int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.priority = p
	c.hasPriority = true
}

func (c *spanContext) samplingPriority() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.priority
}

func (c *spanContext) hasSamplingPriority() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hasPriority
}

func (c *spanContext) setBaggageItem(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.baggage == nil {
		c.baggage = make(map[string]string, 1)
	}
	c.baggage[key] = val
}

func (c *spanContext) baggageItem(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baggage[key]
}

// finish marks this span as finished in the trace.
func (c *spanContext) finish() { c.trace.ackFinish() }

// trace holds information about a specific trace. This structure is shared
// between all spans in a trace.
type trace struct {
	mu       sync.RWMutex // guards below fields
	spans    []*span      // all the spans that are part of this trace
	finished int          // the number of finished spans
	full     bool         // signifies that the span buffer is full
}

var (
	// traceStartSize is the initial size of our trace buffer,
	// by default we allocate for a handful of spans within the trace,
	// reasonable as span is actually way bigger, and avoids re-allocating
	// over and over. Could be fine-tuned at runtime.
	traceStartSize = 10
	// traceMaxSize is the maximum number of spans we keep in memory.
	// This is to avoid memory leaks, if above that value, spans are randomly
	// dropped and ignore, resulting in corrupted tracing data, but ensuring
	// original program continues to work as expected.
	traceMaxSize = int(1e5)
)

// newTrace creates a new trace using the given callback which will be called
// upon completion of the trace.
func newTrace() *trace {
	return &trace{spans: make([]*span, 0, traceStartSize)}
}

// push pushes a new span into the trace. If the buffer is full, it returns
// a errBufferFull error.
func (t *trace) push(sp *span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.full {
		return
	}
	if len(t.spans) >= traceMaxSize {
		// capacity is reached, we will not be able to complete this trace.
		t.full = true
		t.spans = nil // GC
		if tr, ok := internal.GetGlobalTracer().(*tracer); ok {
			// we have a tracer we can submit errors too.
			tr.pushError(&spanBufferFullError{})
		}
		return
	}
	t.spans = append(t.spans, sp)
}

// ackFinish aknowledges that another span in the trace has finished, and checks
// if the trace is complete, in which case it calls the onFinish function.
func (t *trace) ackFinish() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.full {
		// capacity has been reached, the buffer is no longer tracking
		// all the spans in the trace, so the below conditions will not
		// be accurate and would trigger a pre-mature flush, exposing us
		// to a race condition where spans can be modified while flushing.
		return
	}
	t.finished++
	if len(t.spans) != t.finished {
		return
	}
	if tr, ok := internal.GetGlobalTracer().(*tracer); ok {
		// we have a tracer that can receive completed traces.
		tr.pushTrace(t.spans)
	}
	t.spans = nil
	t.finished = 0 // important, because a buffer can be used for several flushes
}
