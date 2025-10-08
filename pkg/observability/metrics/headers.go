package metrics

import (
	"net/http"

	"github.com/go-kit/kit/metrics"
)

// CounterWithHeaders represents a counter that can use http.Header values as label values.
type CounterWithHeaders interface {
	Add(delta float64)
	With(headers http.Header, labelValues ...string) CounterWithHeaders
}

// MultiCounterWithHeaders collects multiple individual CounterWithHeaders and treats them as a unit.
type MultiCounterWithHeaders []CounterWithHeaders

// NewMultiCounterWithHeaders returns a multi-counter, wrapping the passed CounterWithHeaders.
func NewMultiCounterWithHeaders(c ...CounterWithHeaders) MultiCounterWithHeaders {
	return c
}

// Add adds the given delta value to the counter value.
func (c MultiCounterWithHeaders) Add(delta float64) {
	for _, counter := range c {
		counter.Add(delta)
	}
}

// With creates a new counter by appending the given label values and http.Header as labels and returns it.
func (c MultiCounterWithHeaders) With(headers http.Header, labelValues ...string) CounterWithHeaders {
	next := make(MultiCounterWithHeaders, len(c))
	for i := range c {
		next[i] = c[i].With(headers, labelValues...)
	}
	return next
}

// NewCounterWithNoopHeaders returns a CounterWithNoopHeaders.
func NewCounterWithNoopHeaders(counter metrics.Counter) CounterWithNoopHeaders {
	return CounterWithNoopHeaders{counter: counter}
}

// CounterWithNoopHeaders is a counter that satisfies CounterWithHeaders but ignores the given http.Header.
type CounterWithNoopHeaders struct {
	counter metrics.Counter
}

// Add adds the given delta value to the counter value.
func (c CounterWithNoopHeaders) Add(delta float64) {
	c.counter.Add(delta)
}

// With creates a new counter by appending the given label values and returns it.
func (c CounterWithNoopHeaders) With(_ http.Header, labelValues ...string) CounterWithHeaders {
	return NewCounterWithNoopHeaders(c.counter.With(labelValues...))
}
