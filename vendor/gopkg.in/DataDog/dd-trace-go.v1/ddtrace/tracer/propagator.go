package tracer

import (
	"errors"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
)

// Propagator implementations should be able to inject and extract
// SpanContexts into an implementation specific carrier.
type Propagator interface {
	// Inject takes the SpanContext and injects it into the carrier.
	Inject(context ddtrace.SpanContext, carrier interface{}) error

	// Extract returns the SpanContext from the given carrier.
	Extract(carrier interface{}) (ddtrace.SpanContext, error)
}

// TextMapWriter allows setting key/value pairs of strings on the underlying
// data structure. Carriers implementing TextMapWriter are compatible to be
// used with Datadog's TextMapPropagator.
type TextMapWriter interface {
	// Set sets the given key/value pair.
	Set(key, val string)
}

// TextMapReader allows iterating over sets of key/value pairs. Carriers implementing
// TextMapReader are compatible to be used with Datadog's TextMapPropagator.
type TextMapReader interface {
	// ForeachKey iterates over all keys that exist in the underlying
	// carrier. It takes a callback function which will be called
	// using all key/value pairs as arguments. ForeachKey will return
	// the first error returned by the handler.
	ForeachKey(handler func(key, val string) error) error
}

var (
	// ErrInvalidCarrier is returned when the carrier provided to the propagator
	// does not implemented the correct interfaces.
	ErrInvalidCarrier = errors.New("invalid carrier")

	// ErrInvalidSpanContext is returned when the span context found in the
	// carrier is not of the expected type.
	ErrInvalidSpanContext = errors.New("invalid span context")

	// ErrSpanContextCorrupted is returned when there was a problem parsing
	// the information found in the carrier.
	ErrSpanContextCorrupted = errors.New("span context corrupted")

	// ErrSpanContextNotFound represents missing information in the given carrier.
	ErrSpanContextNotFound = errors.New("span context not found")
)
