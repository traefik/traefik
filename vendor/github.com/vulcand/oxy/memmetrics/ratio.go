package memmetrics

import (
	"time"

	"github.com/mailgun/timetools"
)

type ratioOptSetter func(r *RatioCounter) error

// RatioClock sets a clock
func RatioClock(clock timetools.TimeProvider) ratioOptSetter {
	return func(r *RatioCounter) error {
		r.clock = clock
		return nil
	}
}

// RatioCounter calculates a ratio of a/a+b over a rolling window of predefined buckets
type RatioCounter struct {
	clock timetools.TimeProvider
	a     *RollingCounter
	b     *RollingCounter
}

// NewRatioCounter creates a new RatioCounter
func NewRatioCounter(buckets int, resolution time.Duration, options ...ratioOptSetter) (*RatioCounter, error) {
	rc := &RatioCounter{}

	for _, o := range options {
		if err := o(rc); err != nil {
			return nil, err
		}
	}

	if rc.clock == nil {
		rc.clock = &timetools.RealTime{}
	}

	a, err := NewCounter(buckets, resolution, CounterClock(rc.clock))
	if err != nil {
		return nil, err
	}

	b, err := NewCounter(buckets, resolution, CounterClock(rc.clock))
	if err != nil {
		return nil, err
	}

	rc.a = a
	rc.b = b
	return rc, nil
}

// Reset reset the counter
func (r *RatioCounter) Reset() {
	r.a.Reset()
	r.b.Reset()
}

// IsReady returns true if the counter is ready
func (r *RatioCounter) IsReady() bool {
	return r.a.countedBuckets+r.b.countedBuckets >= len(r.a.values)
}

// CountA gets count A
func (r *RatioCounter) CountA() int64 {
	return r.a.Count()
}

// CountB gets count B
func (r *RatioCounter) CountB() int64 {
	return r.b.Count()
}

// Resolution gets resolution
func (r *RatioCounter) Resolution() time.Duration {
	return r.a.Resolution()
}

// Buckets gets buckets
func (r *RatioCounter) Buckets() int {
	return r.a.Buckets()
}

// WindowSize gets windows size
func (r *RatioCounter) WindowSize() time.Duration {
	return r.a.WindowSize()
}

// ProcessedCount gets processed count
func (r *RatioCounter) ProcessedCount() int64 {
	return r.CountA() + r.CountB()
}

// Ratio gets ratio
func (r *RatioCounter) Ratio() float64 {
	a := r.a.Count()
	b := r.b.Count()
	// No data, return ok
	if a+b == 0 {
		return 0
	}
	return float64(a) / float64(a+b)
}

// IncA increment counter A
func (r *RatioCounter) IncA(v int) {
	r.a.Inc(v)
}

// IncB increment counter B
func (r *RatioCounter) IncB(v int) {
	r.b.Inc(v)
}

// TestMeter a test meter
type TestMeter struct {
	Rate       float64
	NotReady   bool
	WindowSize time.Duration
}

// GetWindowSize gets windows size
func (tm *TestMeter) GetWindowSize() time.Duration {
	return tm.WindowSize
}

// IsReady returns true if the meter is ready
func (tm *TestMeter) IsReady() bool {
	return !tm.NotReady
}

// GetRate gets rate
func (tm *TestMeter) GetRate() float64 {
	return tm.Rate
}
