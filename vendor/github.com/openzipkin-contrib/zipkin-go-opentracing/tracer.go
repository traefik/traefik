package zipkintracer

import (
	"errors"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	otobserver "github.com/opentracing-contrib/go-observer"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/flag"
)

// ErrInvalidEndpoint will be thrown if hostPort parameter is corrupted or host
// can't be resolved
var ErrInvalidEndpoint = errors.New("Invalid Endpoint. Please check hostPort parameter")

// Tracer extends the opentracing.Tracer interface with methods to
// probe implementation state, for use by zipkintracer consumers.
type Tracer interface {
	opentracing.Tracer

	// Options gets the Options used in New() or NewWithOptions().
	Options() TracerOptions
}

// TracerOptions allows creating a customized Tracer.
type TracerOptions struct {
	// shouldSample is a function which is called when creating a new Span and
	// determines whether that Span is sampled. The randomized TraceID is supplied
	// to allow deterministic sampling decisions to be made across different nodes.
	shouldSample func(traceID uint64) bool
	// trimUnsampledSpans turns potentially expensive operations on unsampled
	// Spans into no-ops. More precisely, tags and log events are silently
	// discarded. If NewSpanEventListener is set, the callbacks will still fire.
	trimUnsampledSpans bool
	// recorder receives Spans which have been finished.
	recorder SpanRecorder
	// newSpanEventListener can be used to enhance the tracer by effectively
	// attaching external code to trace events. See NetTraceIntegrator for a
	// practical example, and event.go for the list of possible events.
	newSpanEventListener func() func(SpanEvent)
	// dropAllLogs turns log events on all Spans into no-ops.
	// If NewSpanEventListener is set, the callbacks will still fire.
	dropAllLogs bool
	// MaxLogsPerSpan limits the number of Logs in a span (if set to a nonzero
	// value). If a span has more logs than this value, logs are dropped as
	// necessary (and replaced with a log describing how many were dropped).
	//
	// About half of the MaxLogPerSpan logs kept are the oldest logs, and about
	// half are the newest logs.
	//
	// If NewSpanEventListener is set, the callbacks will still fire for all log
	// events. This value is ignored if DropAllLogs is true.
	maxLogsPerSpan int
	// debugAssertSingleGoroutine internally records the ID of the goroutine
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
	debugAssertSingleGoroutine bool
	// debugAssertUseAfterFinish is provided strictly for development purposes.
	// When set, it attempts to exacerbate issues emanating from use of Spans
	// after calling Finish by running additional assertions.
	debugAssertUseAfterFinish bool
	// enableSpanPool enables the use of a pool, so that the tracer reuses spans
	// after Finish has been called on it. Adds a slight performance gain as it
	// reduces allocations. However, if you have any use-after-finish race
	// conditions the code may panic.
	enableSpanPool bool
	// logger ...
	logger Logger
	// clientServerSameSpan allows for Zipkin V1 style span per RPC. This places
	// both client end and server end of a RPC call into the same span.
	clientServerSameSpan bool
	// debugMode activates Zipkin's debug request allowing for all Spans originating
	// from this tracer to pass through and bypass sampling. Use with extreme care
	// as it might flood your system if you have many traces starting from the
	// service you are instrumenting.
	debugMode bool
	// traceID128Bit enables the generation of 128 bit traceIDs in case the tracer
	// needs to create a root span. By default regular 64 bit traceIDs are used.
	// Regardless of this setting, the library will propagate and support both
	// 64 and 128 bit incoming traces from upstream sources.
	traceID128Bit bool

	observer otobserver.Observer
}

// TracerOption allows for functional options.
// See: http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type TracerOption func(opts *TracerOptions) error

// WithSampler allows one to add a Sampler function
func WithSampler(sampler Sampler) TracerOption {
	return func(opts *TracerOptions) error {
		opts.shouldSample = sampler
		return nil
	}
}

// TrimUnsampledSpans option
func TrimUnsampledSpans(trim bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.trimUnsampledSpans = trim
		return nil
	}
}

// DropAllLogs option
func DropAllLogs(dropAllLogs bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.dropAllLogs = dropAllLogs
		return nil
	}
}

// WithLogger option
func WithLogger(logger Logger) TracerOption {
	return func(opts *TracerOptions) error {
		opts.logger = logger
		return nil
	}
}

// DebugAssertSingleGoroutine option
func DebugAssertSingleGoroutine(val bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.debugAssertSingleGoroutine = val
		return nil
	}
}

// DebugAssertUseAfterFinish option
func DebugAssertUseAfterFinish(val bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.debugAssertUseAfterFinish = val
		return nil
	}
}

// TraceID128Bit option
func TraceID128Bit(val bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.traceID128Bit = val
		return nil
	}
}

// ClientServerSameSpan allows to place client-side and server-side annotations
// for a RPC call in the same span (Zipkin V1 behavior) or different spans
// (more in line with other tracing solutions). By default this Tracer
// uses shared host spans (so client-side and server-side in the same span).
// If using separate spans you might run into trouble with Zipkin V1 as clock
// skew issues can't be remedied at Zipkin server side.
func ClientServerSameSpan(val bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.clientServerSameSpan = val
		return nil
	}
}

// DebugMode allows to set the tracer to Zipkin debug mode
func DebugMode(val bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.debugMode = val
		return nil
	}
}

// EnableSpanPool ...
func EnableSpanPool(val bool) TracerOption {
	return func(opts *TracerOptions) error {
		opts.enableSpanPool = val
		return nil
	}
}

// NewSpanEventListener option
func NewSpanEventListener(f func() func(SpanEvent)) TracerOption {
	return func(opts *TracerOptions) error {
		opts.newSpanEventListener = f
		return nil
	}
}

// WithMaxLogsPerSpan option
func WithMaxLogsPerSpan(limit int) TracerOption {
	return func(opts *TracerOptions) error {
		if limit < 5 || limit > 10000 {
			return errors.New("invalid MaxLogsPerSpan limit. Should be between 5 and 10000")
		}
		opts.maxLogsPerSpan = limit
		return nil
	}
}

// NewTracer creates a new OpenTracing compatible Zipkin Tracer.
func NewTracer(recorder SpanRecorder, options ...TracerOption) (opentracing.Tracer, error) {
	opts := &TracerOptions{
		recorder:             recorder,
		shouldSample:         alwaysSample,
		trimUnsampledSpans:   false,
		newSpanEventListener: func() func(SpanEvent) { return nil },
		logger:               &nopLogger{},
		debugAssertSingleGoroutine: false,
		debugAssertUseAfterFinish:  false,
		clientServerSameSpan:       true,
		debugMode:                  false,
		traceID128Bit:              false,
		maxLogsPerSpan:             10000,
		observer:                   nil,
	}
	for _, o := range options {
		err := o(opts)
		if err != nil {
			return nil, err
		}
	}
	rval := &tracerImpl{options: *opts}
	rval.textPropagator = &textMapPropagator{rval}
	rval.binaryPropagator = &binaryPropagator{rval}
	rval.accessorPropagator = &accessorPropagator{rval}
	return rval, nil
}

// Implements the `Tracer` interface.
type tracerImpl struct {
	options            TracerOptions
	textPropagator     *textMapPropagator
	binaryPropagator   *binaryPropagator
	accessorPropagator *accessorPropagator
}

func (t *tracerImpl) StartSpan(
	operationName string,
	opts ...opentracing.StartSpanOption,
) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return t.startSpanWithOptions(operationName, sso)
}

func (t *tracerImpl) getSpan() *spanImpl {
	if t.options.enableSpanPool {
		sp := spanPool.Get().(*spanImpl)
		sp.reset()
		return sp
	}
	return &spanImpl{}
}

func (t *tracerImpl) startSpanWithOptions(
	operationName string,
	opts opentracing.StartSpanOptions,
) opentracing.Span {
	// Start time.
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	// Tags.
	tags := opts.Tags

	// Build the new span. This is the only allocation: We'll return this as
	// an opentracing.Span.
	sp := t.getSpan()

	if t.options.observer != nil {
		sp.observer, _ = t.options.observer.OnStartSpan(sp, operationName, opts)
	}

	// Look for a parent in the list of References.
	//
	// TODO: would be nice if basictracer did something with all
	// References, not just the first one.
ReferencesLoop:
	for _, ref := range opts.References {
		switch ref.Type {
		case opentracing.ChildOfRef:
			refCtx := ref.ReferencedContext.(SpanContext)
			sp.raw.Context.TraceID = refCtx.TraceID
			sp.raw.Context.ParentSpanID = &refCtx.SpanID
			sp.raw.Context.Sampled = refCtx.Sampled
			sp.raw.Context.Flags = refCtx.Flags
			sp.raw.Context.Flags &^= flag.IsRoot // unset IsRoot flag if needed

			if t.options.clientServerSameSpan &&
				tags[string(ext.SpanKind)] == ext.SpanKindRPCServer.Value {
				sp.raw.Context.SpanID = refCtx.SpanID
				sp.raw.Context.ParentSpanID = refCtx.ParentSpanID
				sp.raw.Context.Owner = false
			} else {
				sp.raw.Context.SpanID = randomID()
				sp.raw.Context.ParentSpanID = &refCtx.SpanID
				sp.raw.Context.Owner = true
			}

			if l := len(refCtx.Baggage); l > 0 {
				sp.raw.Context.Baggage = make(map[string]string, l)
				for k, v := range refCtx.Baggage {
					sp.raw.Context.Baggage[k] = v
				}
			}
			break ReferencesLoop
		case opentracing.FollowsFromRef:
			refCtx := ref.ReferencedContext.(SpanContext)
			sp.raw.Context.TraceID = refCtx.TraceID
			sp.raw.Context.ParentSpanID = &refCtx.SpanID
			sp.raw.Context.Sampled = refCtx.Sampled
			sp.raw.Context.Flags = refCtx.Flags
			sp.raw.Context.Flags &^= flag.IsRoot // unset IsRoot flag if needed

			sp.raw.Context.SpanID = randomID()
			sp.raw.Context.ParentSpanID = &refCtx.SpanID
			sp.raw.Context.Owner = true

			if l := len(refCtx.Baggage); l > 0 {
				sp.raw.Context.Baggage = make(map[string]string, l)
				for k, v := range refCtx.Baggage {
					sp.raw.Context.Baggage[k] = v
				}
			}
			break ReferencesLoop
		}
	}
	if sp.raw.Context.TraceID.Empty() {
		// No parent Span found; allocate new trace and span ids and determine
		// the Sampled status.
		if t.options.traceID128Bit {
			sp.raw.Context.TraceID.High = randomID()
		}
		sp.raw.Context.TraceID.Low, sp.raw.Context.SpanID = randomID2()
		sp.raw.Context.Sampled = t.options.shouldSample(sp.raw.Context.TraceID.Low)
		sp.raw.Context.Flags = flag.IsRoot
		sp.raw.Context.Owner = true
	}
	if t.options.debugMode {
		sp.raw.Context.Flags |= flag.Debug
	}
	return t.startSpanInternal(
		sp,
		operationName,
		startTime,
		tags,
	)
}

func (t *tracerImpl) startSpanInternal(
	sp *spanImpl,
	operationName string,
	startTime time.Time,
	tags opentracing.Tags,
) opentracing.Span {
	sp.tracer = t
	if t.options.newSpanEventListener != nil {
		sp.event = t.options.newSpanEventListener()
	}
	sp.raw.Operation = operationName
	sp.raw.Start = startTime
	sp.raw.Duration = -1
	sp.raw.Tags = tags

	if t.options.debugAssertSingleGoroutine {
		sp.SetTag(debugGoroutineIDTag, curGoroutineID())
	}
	defer sp.onCreate(operationName)
	return sp
}

type delegatorType struct{}

// Delegator is the format to use for DelegatingCarrier.
var Delegator delegatorType

func (t *tracerImpl) Inject(sc opentracing.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return t.textPropagator.Inject(sc, carrier)
	case opentracing.Binary:
		return t.binaryPropagator.Inject(sc, carrier)
	}
	if _, ok := format.(delegatorType); ok {
		return t.accessorPropagator.Inject(sc, carrier)
	}
	return opentracing.ErrUnsupportedFormat
}

func (t *tracerImpl) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return t.textPropagator.Extract(carrier)
	case opentracing.Binary:
		return t.binaryPropagator.Extract(carrier)
	}
	if _, ok := format.(delegatorType); ok {
		return t.accessorPropagator.Extract(carrier)
	}
	return nil, opentracing.ErrUnsupportedFormat
}

func (t *tracerImpl) Options() TracerOptions {
	return t.options
}

// WithObserver assigns an initialized observer to opts.observer
func WithObserver(observer otobserver.Observer) TracerOption {
	return func(opts *TracerOptions) error {
		opts.observer = observer
		return nil
	}
}
