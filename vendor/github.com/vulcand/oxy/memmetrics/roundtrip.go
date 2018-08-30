package memmetrics

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/mailgun/timetools"
)

// RTMetrics provides aggregated performance metrics for HTTP requests processing
// such as round trip latency, response codes counters network error and total requests.
// all counters are collected as rolling window counters with defined precision, histograms
// are a rolling window histograms with defined precision as well.
// See RTOptions for more detail on parameters.
type RTMetrics struct {
	total           *RollingCounter
	netErrors       *RollingCounter
	statusCodes     map[int]*RollingCounter
	statusCodesLock sync.RWMutex
	histogram       *RollingHDRHistogram
	histogramLock   sync.RWMutex

	newCounter NewCounterFn
	newHist    NewRollingHistogramFn
	clock      timetools.TimeProvider
}

type rrOptSetter func(r *RTMetrics) error

// NewRTMetricsFn builder function type
type NewRTMetricsFn func() (*RTMetrics, error)

// NewCounterFn builder function type
type NewCounterFn func() (*RollingCounter, error)

// NewRollingHistogramFn builder function type
type NewRollingHistogramFn func() (*RollingHDRHistogram, error)

// RTCounter set a builder function for Counter
func RTCounter(new NewCounterFn) rrOptSetter {
	return func(r *RTMetrics) error {
		r.newCounter = new
		return nil
	}
}

// RTHistogram set a builder function for RollingHistogram
func RTHistogram(fn NewRollingHistogramFn) rrOptSetter {
	return func(r *RTMetrics) error {
		r.newHist = fn
		return nil
	}
}

// RTClock sets a clock
func RTClock(clock timetools.TimeProvider) rrOptSetter {
	return func(r *RTMetrics) error {
		r.clock = clock
		return nil
	}
}

// NewRTMetrics returns new instance of metrics collector.
func NewRTMetrics(settings ...rrOptSetter) (*RTMetrics, error) {
	m := &RTMetrics{
		statusCodes:     make(map[int]*RollingCounter),
		statusCodesLock: sync.RWMutex{},
	}
	for _, s := range settings {
		if err := s(m); err != nil {
			return nil, err
		}
	}

	if m.clock == nil {
		m.clock = &timetools.RealTime{}
	}

	if m.newCounter == nil {
		m.newCounter = func() (*RollingCounter, error) {
			return NewCounter(counterBuckets, counterResolution, CounterClock(m.clock))
		}
	}

	if m.newHist == nil {
		m.newHist = func() (*RollingHDRHistogram, error) {
			return NewRollingHDRHistogram(histMin, histMax, histSignificantFigures, histPeriod, histBuckets, RollingClock(m.clock))
		}
	}

	h, err := m.newHist()
	if err != nil {
		return nil, err
	}

	netErrors, err := m.newCounter()
	if err != nil {
		return nil, err
	}

	total, err := m.newCounter()
	if err != nil {
		return nil, err
	}

	m.histogram = h
	m.netErrors = netErrors
	m.total = total
	return m, nil
}

// Export Returns a new RTMetrics which is a copy of the current one
func (m *RTMetrics) Export() *RTMetrics {
	m.statusCodesLock.RLock()
	defer m.statusCodesLock.RUnlock()
	m.histogramLock.RLock()
	defer m.histogramLock.RUnlock()

	export := &RTMetrics{}
	export.statusCodesLock = sync.RWMutex{}
	export.histogramLock = sync.RWMutex{}
	export.total = m.total.Clone()
	export.netErrors = m.netErrors.Clone()
	exportStatusCodes := map[int]*RollingCounter{}
	for code, rollingCounter := range m.statusCodes {
		exportStatusCodes[code] = rollingCounter.Clone()
	}
	export.statusCodes = exportStatusCodes
	if m.histogram != nil {
		export.histogram = m.histogram.Export()
	}
	export.newCounter = m.newCounter
	export.newHist = m.newHist
	export.clock = m.clock

	return export
}

// CounterWindowSize gets total windows size
func (m *RTMetrics) CounterWindowSize() time.Duration {
	return m.total.WindowSize()
}

// NetworkErrorRatio calculates the amont of network errors such as time outs and dropped connection
// that occurred in the given time window compared to the total requests count.
func (m *RTMetrics) NetworkErrorRatio() float64 {
	if m.total.Count() == 0 {
		return 0
	}
	return float64(m.netErrors.Count()) / float64(m.total.Count())
}

// ResponseCodeRatio calculates ratio of count(startA to endA) / count(startB to endB)
func (m *RTMetrics) ResponseCodeRatio(startA, endA, startB, endB int) float64 {
	a := int64(0)
	b := int64(0)
	m.statusCodesLock.RLock()
	defer m.statusCodesLock.RUnlock()
	for code, v := range m.statusCodes {
		if code < endA && code >= startA {
			a += v.Count()
		}
		if code < endB && code >= startB {
			b += v.Count()
		}
	}
	if b != 0 {
		return float64(a) / float64(b)
	}
	return 0
}

// Append append a metric
func (m *RTMetrics) Append(other *RTMetrics) error {
	if m == other {
		return errors.New("RTMetrics cannot append to self")
	}

	if err := m.total.Append(other.total); err != nil {
		return err
	}

	if err := m.netErrors.Append(other.netErrors); err != nil {
		return err
	}

	copied := other.Export()

	m.statusCodesLock.Lock()
	defer m.statusCodesLock.Unlock()
	m.histogramLock.Lock()
	defer m.histogramLock.Unlock()
	for code, c := range copied.statusCodes {
		o, ok := m.statusCodes[code]
		if ok {
			if err := o.Append(c); err != nil {
				return err
			}
		} else {
			m.statusCodes[code] = c.Clone()
		}
	}

	return m.histogram.Append(copied.histogram)
}

// Record records a metric
func (m *RTMetrics) Record(code int, duration time.Duration) {
	m.total.Inc(1)
	if code == http.StatusGatewayTimeout || code == http.StatusBadGateway {
		m.netErrors.Inc(1)
	}
	m.recordStatusCode(code)
	m.recordLatency(duration)
}

// TotalCount returns total count of processed requests collected.
func (m *RTMetrics) TotalCount() int64 {
	return m.total.Count()
}

// NetworkErrorCount returns total count of processed requests observed
func (m *RTMetrics) NetworkErrorCount() int64 {
	return m.netErrors.Count()
}

// StatusCodesCounts returns map with counts of the response codes
func (m *RTMetrics) StatusCodesCounts() map[int]int64 {
	sc := make(map[int]int64)
	m.statusCodesLock.RLock()
	defer m.statusCodesLock.RUnlock()
	for k, v := range m.statusCodes {
		if v.Count() != 0 {
			sc[k] = v.Count()
		}
	}
	return sc
}

// LatencyHistogram computes and returns resulting histogram with latencies observed.
func (m *RTMetrics) LatencyHistogram() (*HDRHistogram, error) {
	m.histogramLock.Lock()
	defer m.histogramLock.Unlock()
	return m.histogram.Merged()
}

// Reset reset metrics
func (m *RTMetrics) Reset() {
	m.statusCodesLock.Lock()
	defer m.statusCodesLock.Unlock()
	m.histogramLock.Lock()
	defer m.histogramLock.Unlock()
	m.histogram.Reset()
	m.total.Reset()
	m.netErrors.Reset()
	m.statusCodes = make(map[int]*RollingCounter)
}

func (m *RTMetrics) recordLatency(d time.Duration) error {
	m.histogramLock.Lock()
	defer m.histogramLock.Unlock()
	return m.histogram.RecordLatencies(d, 1)
}

func (m *RTMetrics) recordStatusCode(statusCode int) error {
	m.statusCodesLock.Lock()
	if c, ok := m.statusCodes[statusCode]; ok {
		c.Inc(1)
		m.statusCodesLock.Unlock()
		return nil
	}
	m.statusCodesLock.Unlock()

	m.statusCodesLock.Lock()
	defer m.statusCodesLock.Unlock()

	// Check if another goroutine has written our counter already
	if c, ok := m.statusCodes[statusCode]; ok {
		c.Inc(1)
		return nil
	}

	c, err := m.newCounter()
	if err != nil {
		return err
	}
	c.Inc(1)
	m.statusCodes[statusCode] = c
	return nil
}

const (
	counterBuckets         = 10
	counterResolution      = time.Second
	histMin                = 1
	histMax                = 3600000000       // 1 hour in microseconds
	histSignificantFigures = 2                // significant figures (1% precision)
	histBuckets            = 6                // number of sub-histograms in a rolling histogram
	histPeriod             = 10 * time.Second // roll time
)
