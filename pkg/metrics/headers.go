package metrics

import (
	"errors"
	"net/http"
	"time"

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

// HistogramWithHeaders represents a histogram that can use http.Header values as label values.
type HistogramWithHeaders interface {
	Observe(value float64)
	With(headers http.Header, labelValues ...string) HistogramWithHeaders
}

// ScalableHistogramWithHeaders is a Histogram with a predefined time unit, and can use http.Header values as label values.
// used when producing observations without explicitly setting the observed value.
type ScalableHistogramWithHeaders interface {
	With(headers http.Header, labelValues ...string) ScalableHistogramWithHeaders
	Observe(v float64)
	ObserveFromStart(start time.Time)
}

// HistogramWithScaleAndHeaders is a histogram that will convert its observed value to the specified unit.
type HistogramWithScaleAndHeaders struct {
	histogram HistogramWithHeaders
	unit      time.Duration
}

// With implements ScalableHistogramWithHeaders.
func (s *HistogramWithScaleAndHeaders) With(headers http.Header, labelValues ...string) ScalableHistogramWithHeaders {
	h, _ := NewScalableHistogramWithHeaders(s.histogram.With(headers, labelValues...), s.unit)
	return h
}

// ObserveFromStart implements ScalableHistogramWithHeaders.
func (s *HistogramWithScaleAndHeaders) ObserveFromStart(start time.Time) {
	if s.unit <= 0 {
		return
	}

	d := float64(time.Since(start).Nanoseconds()) / float64(s.unit)
	if d < 0 {
		d = 0
	}
	s.histogram.Observe(d)
}

// Observe implements ScalableHistogramWithHeaders.
func (s *HistogramWithScaleAndHeaders) Observe(v float64) {
	s.histogram.Observe(v)
}

// NewScalableHistogramWithHeaders returns a ScalableHistogramWithHeaders. It returns an error if the given unit is <= 0.
func NewScalableHistogramWithHeaders(histogram HistogramWithHeaders, unit time.Duration) (ScalableHistogramWithHeaders, error) {
	if unit <= 0 {
		return nil, errors.New("invalid time unit")
	}
	return &HistogramWithScaleAndHeaders{
		histogram: histogram,
		unit:      unit,
	}, nil
}

// MultiScalableHistogramWithHeaders collects multiple individual histograms and treats them as a unit.
type MultiScalableHistogramWithHeaders []ScalableHistogramWithHeaders

// ObserveFromStart implements ScalableHistogram.
func (h MultiScalableHistogramWithHeaders) ObserveFromStart(start time.Time) {
	for _, histogram := range h {
		histogram.ObserveFromStart(start)
	}
}

// Observe implements ScalableHistogram.
func (h MultiScalableHistogramWithHeaders) Observe(v float64) {
	for _, histogram := range h {
		histogram.Observe(v)
	}
}

// With implements ScalableHistogram.
func (h MultiScalableHistogramWithHeaders) With(headers http.Header, labelValues ...string) ScalableHistogramWithHeaders {
	next := make(MultiScalableHistogramWithHeaders, len(h))
	for i := range h {
		next[i] = h[i].With(headers, labelValues...)
	}
	return next
}

// NewMultiScalableHistogramWithHeaders returns a multi-scalable-histogram, wrapping the passed ScalableHistogramWithHeaders.
func NewMultiScalableHistogramWithHeaders(h ...ScalableHistogramWithHeaders) MultiScalableHistogramWithHeaders {
	return h
}

// NewScalableHistogramWithNoopHeaders returns a ScalableHistogramWithNoopHeaders.
func NewScalableHistogramWithNoopHeaders(histogram metrics.Histogram, unit time.Duration) (ScalableHistogramWithHeaders, error) {
	if unit <= 0 {
		return nil, errors.New("invalid time unit")
	}
	return ScalableHistogramWithNoopHeaders{histogram: &HistogramWithScale{
		histogram: histogram,
		unit:      unit,
	}}, nil
}

// NewScalableHistogramWithNoopHeaders returns a ScalableHistogramWithNoopHeaders.
func newScalableHistogramWithNoopHeaders(histogram ScalableHistogram) ScalableHistogramWithHeaders {
	return ScalableHistogramWithNoopHeaders{histogram: histogram}
}

// ScalableHistogramWithNoopHeaders is a histogram that satisfies ScalableHistogramWithHeaders but ignores the given http.Header.
type ScalableHistogramWithNoopHeaders struct {
	histogram ScalableHistogram
}

// Observe adds the given delta value to the histogram value.
func (h ScalableHistogramWithNoopHeaders) Observe(delta float64) {
	h.histogram.Observe(delta)
}

// With creates a new histogram by appending the given label values and returns it.
func (h ScalableHistogramWithNoopHeaders) With(_ http.Header, labelValues ...string) ScalableHistogramWithHeaders {
	return newScalableHistogramWithNoopHeaders(h.histogram.With(labelValues...))
}

// ObserveFromStart adds the given delta value to the histogram value.
func (h ScalableHistogramWithNoopHeaders) ObserveFromStart(start time.Time) {
	h.histogram.ObserveFromStart(start)
}
