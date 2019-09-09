package tracer

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/internal"
)

var _ ddtrace.Tracer = (*tracer)(nil)

// tracer creates, buffers and submits Spans which are used to time blocks of
// computation. They are accumulated and streamed into an internal payload,
// which is flushed to the agent whenever its size exceeds a specific threshold
// or when a certain interval of time has passed, whichever happens first.
//
// tracer operates based on a worker loop which responds to various request
// channels. It additionally holds two buffers which accumulates error and trace
// queues to be processed by the payload encoder.
type tracer struct {
	*config
	*payload

	flushAllReq    chan chan<- struct{}
	flushTracesReq chan struct{}
	flushErrorsReq chan struct{}
	exitReq        chan struct{}

	payloadQueue chan []*span
	errorBuffer  chan error

	// stopped is a channel that will be closed when the worker has exited.
	stopped chan struct{}

	// syncPush is used for testing. When non-nil, it causes pushTrace to become
	// a synchronous (blocking) operation, meaning that it will only return after
	// the trace has been fully processed and added onto the payload.
	syncPush chan struct{}

	// prioritySampling holds an instance of the priority sampler.
	prioritySampling *prioritySampler
}

const (
	// flushInterval is the interval at which the payload contents will be flushed
	// to the transport.
	flushInterval = 2 * time.Second

	// payloadMaxLimit is the maximum payload size allowed and should indicate the
	// maximum size of the package that the agent can receive.
	payloadMaxLimit = 9.5 * 1024 * 1024 // 9.5 MB

	// payloadSizeLimit specifies the maximum allowed size of the payload before
	// it will trigger a flush to the transport.
	payloadSizeLimit = payloadMaxLimit / 2
)

// Start starts the tracer with the given set of options. It will stop and replace
// any running tracer, meaning that calling it several times will result in a restart
// of the tracer by replacing the current instance with a new one.
func Start(opts ...StartOption) {
	if internal.Testing {
		return // mock tracer active
	}
	internal.SetGlobalTracer(newTracer(opts...))
}

// Stop stops the started tracer. Subsequent calls are valid but become no-op.
func Stop() {
	internal.SetGlobalTracer(&internal.NoopTracer{})
}

// Span is an alias for ddtrace.Span. It is here to allow godoc to group methods returning
// ddtrace.Span. It is recommended and is considered more correct to refer to this type as
// ddtrace.Span instead.
type Span = ddtrace.Span

// StartSpan starts a new span with the given operation name and set of options.
// If the tracer is not started, calling this function is a no-op.
func StartSpan(operationName string, opts ...StartSpanOption) Span {
	return internal.GetGlobalTracer().StartSpan(operationName, opts...)
}

// Extract extracts a SpanContext from the carrier. The carrier is expected
// to implement TextMapReader, otherwise an error is returned.
// If the tracer is not started, calling this function is a no-op.
func Extract(carrier interface{}) (ddtrace.SpanContext, error) {
	return internal.GetGlobalTracer().Extract(carrier)
}

// Inject injects the given SpanContext into the carrier. The carrier is
// expected to implement TextMapWriter, otherwise an error is returned.
// If the tracer is not started, calling this function is a no-op.
func Inject(ctx ddtrace.SpanContext, carrier interface{}) error {
	return internal.GetGlobalTracer().Inject(ctx, carrier)
}

const (
	// payloadQueueSize is the buffer size of the trace channel.
	payloadQueueSize = 1000

	// errorBufferSize is the buffer size of the error channel.
	errorBufferSize = 200
)

func newTracer(opts ...StartOption) *tracer {
	c := new(config)
	defaults(c)
	for _, fn := range opts {
		fn(c)
	}
	if c.transport == nil {
		c.transport = newTransport(c.agentAddr, c.httpRoundTripper)
	}
	if c.propagator == nil {
		c.propagator = NewPropagator(nil)
	}
	t := &tracer{
		config:           c,
		payload:          newPayload(),
		flushAllReq:      make(chan chan<- struct{}),
		flushTracesReq:   make(chan struct{}, 1),
		flushErrorsReq:   make(chan struct{}, 1),
		exitReq:          make(chan struct{}),
		payloadQueue:     make(chan []*span, payloadQueueSize),
		errorBuffer:      make(chan error, errorBufferSize),
		stopped:          make(chan struct{}),
		prioritySampling: newPrioritySampler(),
	}

	go t.worker()

	return t
}

// worker receives finished traces to be added into the payload, as well
// as periodically flushes traces to the transport.
func (t *tracer) worker() {
	defer close(t.stopped)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case trace := <-t.payloadQueue:
			t.pushPayload(trace)

		case <-ticker.C:
			t.flush()

		case done := <-t.flushAllReq:
			t.flush()
			done <- struct{}{}

		case <-t.flushTracesReq:
			t.flushTraces()

		case <-t.flushErrorsReq:
			t.flushErrors()

		case <-t.exitReq:
			t.flush()
			return
		}
	}
}

func (t *tracer) pushTrace(trace []*span) {
	select {
	case <-t.stopped:
		return
	default:
	}
	select {
	case t.payloadQueue <- trace:
	default:
		t.pushError(&dataLossError{
			context: errors.New("payload queue full, dropping trace"),
			count:   len(trace),
		})
	}
	if t.syncPush != nil {
		// only in tests
		<-t.syncPush
	}
}

func (t *tracer) pushError(err error) {
	select {
	case <-t.stopped:
		return
	default:
	}
	if len(t.errorBuffer) >= cap(t.errorBuffer)/2 { // starts being full, anticipate, try and flush soon
		select {
		case t.flushErrorsReq <- struct{}{}:
		default: // a flush was already requested, skip
		}
	}
	select {
	case t.errorBuffer <- err:
	default:
		// OK, if we get this, our error error buffer is full,
		// we can assume it is filled with meaningful messages which
		// are going to be logged and hopefully read, nothing better
		// we can do, blocking would make things worse.
	}
}

// StartSpan creates, starts, and returns a new Span with the given `operationName`.
func (t *tracer) StartSpan(operationName string, options ...ddtrace.StartSpanOption) ddtrace.Span {
	var opts ddtrace.StartSpanConfig
	for _, fn := range options {
		fn(&opts)
	}
	var startTime int64
	if opts.StartTime.IsZero() {
		startTime = now()
	} else {
		startTime = opts.StartTime.UnixNano()
	}
	var context *spanContext
	if opts.Parent != nil {
		if ctx, ok := opts.Parent.(*spanContext); ok {
			context = ctx
		}
	}
	id := opts.SpanID
	if id == 0 {
		id = random.Uint64()
	}
	// span defaults
	span := &span{
		Name:     operationName,
		Service:  t.config.serviceName,
		Resource: operationName,
		Meta:     map[string]string{},
		Metrics:  map[string]float64{},
		SpanID:   id,
		TraceID:  id,
		ParentID: 0,
		Start:    startTime,
	}
	if context != nil {
		// this is a child span
		span.TraceID = context.traceID
		span.ParentID = context.spanID
		if context.hasSamplingPriority() {
			span.Metrics[keySamplingPriority] = float64(context.samplingPriority())
		}
		if context.span != nil {
			// local parent, inherit service
			context.span.RLock()
			span.Service = context.span.Service
			context.span.RUnlock()
		} else {
			// remote parent
			if context.origin != "" {
				// mark origin
				span.Meta[keyOrigin] = context.origin
			}
		}
	}
	span.context = newSpanContext(span, context)
	if context == nil || context.span == nil {
		// this is either a root span or it has a remote parent, we should add the PID.
		span.SetTag(ext.Pid, strconv.Itoa(os.Getpid()))
	}
	// add tags from options
	for k, v := range opts.Tags {
		span.SetTag(k, v)
	}
	// add global tags
	for k, v := range t.config.globalTags {
		span.SetTag(k, v)
	}
	if context == nil {
		// this is a brand new trace, sample it
		t.sample(span)
	}
	return span
}

// Stop stops the tracer.
func (t *tracer) Stop() {
	select {
	case <-t.stopped:
		return
	default:
		t.exitReq <- struct{}{}
		<-t.stopped
	}
}

// Inject uses the configured or default TextMap Propagator.
func (t *tracer) Inject(ctx ddtrace.SpanContext, carrier interface{}) error {
	return t.config.propagator.Inject(ctx, carrier)
}

// Extract uses the configured or default TextMap Propagator.
func (t *tracer) Extract(carrier interface{}) (ddtrace.SpanContext, error) {
	return t.config.propagator.Extract(carrier)
}

// flushTraces will push any currently buffered traces to the server.
func (t *tracer) flushTraces() {
	if t.payload.itemCount() == 0 {
		return
	}
	size, count := t.payload.size(), t.payload.itemCount()
	if t.config.debug {
		log.Printf("Sending payload: size: %d traces: %d\n", size, count)
	}
	rc, err := t.config.transport.send(t.payload)
	if err != nil {
		t.pushError(&dataLossError{context: err, count: count})
	}
	if err == nil {
		t.prioritySampling.readRatesJSON(rc) // TODO: handle error?
	}
	t.payload.reset()
}

// flushErrors will process log messages that were queued
func (t *tracer) flushErrors() {
	logErrors(t.errorBuffer)
}

func (t *tracer) flush() {
	t.flushTraces()
	t.flushErrors()
}

// forceFlush forces a flush of data (traces and services) to the agent.
// Flushes are done by a background task on a regular basis, so you never
// need to call this manually, mostly useful for testing and debugging.
func (t *tracer) forceFlush() {
	done := make(chan struct{})
	t.flushAllReq <- done
	<-done
}

// pushPayload pushes the trace onto the payload. If the payload becomes
// larger than the threshold as a result, it sends a flush request.
func (t *tracer) pushPayload(trace []*span) {
	if err := t.payload.push(trace); err != nil {
		t.pushError(&traceEncodingError{context: err})
	}
	if t.payload.size() > payloadSizeLimit {
		// getting large
		select {
		case t.flushTracesReq <- struct{}{}:
		default:
			// flush already queued
		}
	}
	if t.syncPush != nil {
		// only in tests
		t.syncPush <- struct{}{}
	}
}

// sampleRateMetricKey is the metric key holding the applied sample rate. Has to be the same as the Agent.
const sampleRateMetricKey = "_sample_rate"

// Sample samples a span with the internal sampler.
func (t *tracer) sample(span *span) {
	if span.context.hasSamplingPriority() {
		// sampling decision was already made
		return
	}
	sampler := t.config.sampler
	if !sampler.Sample(span) {
		span.context.drop = true
		return
	}
	if rs, ok := sampler.(RateSampler); ok && rs.Rate() < 1 {
		span.Metrics[sampleRateMetricKey] = rs.Rate()
	}
	t.prioritySampling.apply(span)
}
