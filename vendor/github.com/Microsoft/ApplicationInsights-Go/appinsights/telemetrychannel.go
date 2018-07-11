package appinsights

import "time"

// Implementations of TelemetryChannel are responsible for queueing and
// periodically submitting telemetry items.
type TelemetryChannel interface {
	// The address of the endpoint to which telemetry is sent
	EndpointAddress() string

	// Queues a single telemetry item
	Send(Telemetry)

	// Forces the current queue to be sent
	Flush()

	// Tears down the submission goroutines, closes internal channels.
	// Any telemetry waiting to be sent is discarded.  Further calls to
	// Send() have undefined behavior.  This is a more abrupt version of
	// Close().
	Stop()

	// Returns true if this channel has been throttled by the data
	// collector.
	IsThrottled() bool

	// Flushes and tears down the submission goroutine and closes
	// internal channels.  Returns a channel that is closed when all
	// pending telemetry items have been submitted and it is safe to
	// shut down without losing telemetry.
	//
	// If retryTimeout is specified and non-zero, then failed
	// submissions will be retried until one succeeds or the timeout
	// expires, whichever occurs first.  A retryTimeout of zero
	// indicates that failed submissions will be retried as usual.  An
	// omitted retryTimeout indicates that submissions should not be
	// retried if they fail.
	//
	// Note that the returned channel may not be closed before
	// retryTimeout even if it is specified.  This is because
	// retryTimeout only applies to the latest telemetry buffer.  This
	// may be typical for applications that submit a large amount of
	// telemetry or are prone to being throttled.  When exiting, you
	// should select on the result channel and your own timer to avoid
	// long delays.
	Close(retryTimeout ...time.Duration) <-chan struct{}
}
