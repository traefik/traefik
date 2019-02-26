package instana

import (
	bt "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

// Tracer extends the opentracing.Tracer interface
type Tracer interface {
	opentracing.Tracer

	// Options gets the Options used in New() or NewWithOptions().
	Options() TracerOptions
}

// TracerOptions allows creating a customized Tracer via NewWithOptions. The object
// must not be updated when there is an active tracer using it.
type TracerOptions struct {
	// ShouldSample is a function which is called when creating a new Span and
	// determines whether that Span is sampled. The randomized TraceID is supplied
	// to allow deterministic sampling decisions to be made across different nodes.
	// For example,
	//
	//   func(traceID uint64) { return traceID % 64 == 0 }
	//
	// samples every 64th trace on average.
	ShouldSample func(traceID int64) bool
	// TrimUnsampledSpans turns potentially expensive operations on unsampled
	// Spans into no-ops. More precisely, tags and log events are silently
	// discarded. If NewSpanEventListener is set, the callbacks will still fire.
	TrimUnsampledSpans bool
	// Recorder receives Spans which have been finished.
	Recorder SpanRecorder
	// NewSpanEventListener can be used to enhance the tracer by effectively
	// attaching external code to trace events. See NetTraceIntegrator for a
	// practical example, and event.go for the list of possible events.
	NewSpanEventListener func() func(bt.SpanEvent)
	// DropAllLogs turns log events on all Spans into no-ops.
	// If NewSpanEventListener is set, the callbacks will still fire.
	DropAllLogs bool
	// MaxLogsPerSpan limits the number of Logs in a span (if set to a nonzero
	// value). If a span has more logs than this value, logs are dropped as
	// necessary (and replaced with a log describing how many were dropped).
	//
	// About half of the MaxLogPerSpan logs kept are the oldest logs, and about
	// half are the newest logs.
	//
	// If NewSpanEventListener is set, the callbacks will still fire for all log
	// events. This value is ignored if DropAllLogs is true.
	MaxLogsPerSpan int
	// DebugAssertSingleGoroutine internally records the ID of the goroutine
	// creating each Span and verifies that no operation is carried out on
	// it on a different goroutine.
	// Provided strictly for development purposes.
	// Passing Spans between goroutine without proper synchronization often
	// results in use-after-Finish() errors. For a simple example, consider the
	// following pseudocode:
	//
	//  func (s *Server) Handle(req http.Request) error {
	//    sp := s.StartSpan("server")
	//    defer sp.Finish()
	//    wait := s.queueProcessing(opentracing.ContextWithSpan(context.Background(), sp), req)
	//    select {
	//    case resp := <-wait:
	//      return resp.Error
	//    case <-time.After(10*time.Second):
	//      sp.LogEvent("timed out waiting for processing")
	//      return ErrTimedOut
	//    }
	//  }
	//
	// This looks reasonable at first, but a request which spends more than ten
	// seconds in the queue is abandoned by the main goroutine and its trace
	// finished, leading to use-after-finish when the request is finally
	// processed. Note also that even joining on to a finished Span via
	// StartSpanWithOptions constitutes an illegal operation.
	//
	// Code bases which do not require (or decide they do not want) Spans to
	// be passed across goroutine boundaries can run with this flag enabled in
	// tests to increase their chances of spotting wrong-doers.
	DebugAssertSingleGoroutine bool
	// DebugAssertUseAfterFinish is provided strictly for development purposes.
	// When set, it attempts to exacerbate issues emanating from use of Spans
	// after calling Finish by running additional assertions.
	DebugAssertUseAfterFinish bool
	// EnableSpanPool enables the use of a pool, so that the tracer reuses spans
	// after Finish has been called on it. Adds a slight performance gain as it
	// reduces allocations. However, if you have any use-after-finish race
	// conditions the code may panic.
	EnableSpanPool bool
}
