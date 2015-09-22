package memmetrics

import (
	"net/http"
	"time"

	"github.com/mailgun/timetools"
)

// RTMetrics provides aggregated performance metrics for HTTP requests processing
// such as round trip latency, response codes counters network error and total requests.
// all counters are collected as rolling window counters with defined precision, histograms
// are a rolling window histograms with defined precision as well.
// See RTOptions for more detail on parameters.
type RTMetrics struct {
	total       *RollingCounter
	netErrors   *RollingCounter
	statusCodes map[int]*RollingCounter
	histogram   *RollingHDRHistogram

	newCounter NewCounterFn
	newHist    NewRollingHistogramFn
	clock      timetools.TimeProvider
}

type rrOptSetter func(r *RTMetrics) error

type NewRTMetricsFn func() (*RTMetrics, error)
type NewCounterFn func() (*RollingCounter, error)
type NewRollingHistogramFn func() (*RollingHDRHistogram, error)

func RTCounter(new NewCounterFn) rrOptSetter {
	return func(r *RTMetrics) error {
		r.newCounter = new
		return nil
	}
}

func RTHistogram(new NewRollingHistogramFn) rrOptSetter {
	return func(r *RTMetrics) error {
		r.newHist = new
		return nil
	}
}

func RTClock(clock timetools.TimeProvider) rrOptSetter {
	return func(r *RTMetrics) error {
		r.clock = clock
		return nil
	}
}

// NewRTMetrics returns new instance of metrics collector.
func NewRTMetrics(settings ...rrOptSetter) (*RTMetrics, error) {
	m := &RTMetrics{
		statusCodes: make(map[int]*RollingCounter),
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

func (m *RTMetrics) CounterWindowSize() time.Duration {
	return m.total.WindowSize()
}

// GetNetworkErrorRatio calculates the amont of network errors such as time outs and dropped connection
// that occured in the given time window compared to the total requests count.
func (m *RTMetrics) NetworkErrorRatio() float64 {
	if m.total.Count() == 0 {
		return 0
	}
	return float64(m.netErrors.Count()) / float64(m.total.Count())
}

// GetResponseCodeRatio calculates ratio of count(startA to endA) / count(startB to endB)
func (m *RTMetrics) ResponseCodeRatio(startA, endA, startB, endB int) float64 {
	a := int64(0)
	b := int64(0)
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

func (m *RTMetrics) Append(other *RTMetrics) error {
	if err := m.total.Append(other.total); err != nil {
		return err
	}

	if err := m.netErrors.Append(other.netErrors); err != nil {
		return err
	}

	for code, c := range other.statusCodes {
		o, ok := m.statusCodes[code]
		if ok {
			if err := o.Append(c); err != nil {
				return err
			}
		} else {
			m.statusCodes[code] = c.Clone()
		}
	}

	return m.histogram.Append(other.histogram)
}

func (m *RTMetrics) Record(code int, duration time.Duration) {
	m.total.Inc(1)
	if code == http.StatusGatewayTimeout || code == http.StatusBadGateway {
		m.netErrors.Inc(1)
	}
	m.recordStatusCode(code)
	m.recordLatency(duration)
}

// GetTotalCount returns total count of processed requests collected.
func (m *RTMetrics) TotalCount() int64 {
	return m.total.Count()
}

// GetNetworkErrorCount returns total count of processed requests observed
func (m *RTMetrics) NetworkErrorCount() int64 {
	return m.netErrors.Count()
}

// GetStatusCodesCounts returns map with counts of the response codes
func (m *RTMetrics) StatusCodesCounts() map[int]int64 {
	sc := make(map[int]int64)
	for k, v := range m.statusCodes {
		if v.Count() != 0 {
			sc[k] = v.Count()
		}
	}
	return sc
}

// GetLatencyHistogram computes and returns resulting histogram with latencies observed.
func (m *RTMetrics) LatencyHistogram() (*HDRHistogram, error) {
	return m.histogram.Merged()
}

func (m *RTMetrics) Reset() {
	m.histogram.Reset()
	m.total.Reset()
	m.netErrors.Reset()
	m.statusCodes = make(map[int]*RollingCounter)
}

func (m *RTMetrics) recordNetError() error {
	m.netErrors.Inc(1)
	return nil
}

func (m *RTMetrics) recordLatency(d time.Duration) error {
	return m.histogram.RecordLatencies(d, 1)
}

func (m *RTMetrics) recordStatusCode(statusCode int) error {
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
	histSignificantFigures = 2                // signigicant figures (1% precision)
	histBuckets            = 6                // number of sub-histograms in a rolling histogram
	histPeriod             = 10 * time.Second // roll time
)
