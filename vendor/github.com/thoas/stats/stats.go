package stats

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// Stats data structure
type Stats struct {
	mu                  sync.RWMutex
	closed              chan struct{}
	Hostname            string
	Uptime              time.Time
	Pid                 int
	ResponseCounts      map[string]int
	TotalResponseCounts map[string]int
	TotalResponseTime   time.Time
	TotalResponseSize   int64
	MetricsCounts       map[string]int
	MetricsTimers       map[string]time.Time
}

// Label data structure
type Label struct {
	Name  string
	Value string
}

// New constructs a new Stats structure
func New() *Stats {
	name, _ := os.Hostname()

	stats := &Stats{
		closed:              make(chan struct{}, 1),
		Uptime:              time.Now(),
		Pid:                 os.Getpid(),
		ResponseCounts:      map[string]int{},
		TotalResponseCounts: map[string]int{},
		TotalResponseTime:   time.Time{},
		Hostname:            name,
	}

	go func() {
		for {
			select {
			case <-stats.closed:
				return
			default:
				stats.ResetResponseCounts()

				time.Sleep(time.Second * 1)
			}
		}
	}()

	return stats
}

func (mw *Stats) Close() {
	close(mw.closed)
}

// ResetResponseCounts reset the response counts
func (mw *Stats) ResetResponseCounts() {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.ResponseCounts = map[string]int{}
}

// Handler is a MiddlewareFunc makes Stats implement the Middleware interface.
func (mw *Stats) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beginning, recorder := mw.Begin(w)

		h.ServeHTTP(recorder, r)

		mw.End(beginning, WithRecorder(recorder))
	})
}

// Negroni compatible interface
func (mw *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	beginning, recorder := mw.Begin(w)

	next(recorder, r)

	mw.End(beginning, WithRecorder(recorder))
}

// Begin starts a recorder
func (mw *Stats) Begin(w http.ResponseWriter) (time.Time, ResponseWriter) {
	start := time.Now()

	writer := NewRecorderResponseWriter(w, 200)

	return start, writer
}

// End closes the recorder with a specific status
func (mw *Stats) End(start time.Time, opts ...Option) {
	options := newOptions(opts...)

	responseTime := time.Since(start)

	mw.mu.Lock()

	defer mw.mu.Unlock()

	// If Hijacked connection do not count in response time
	if options.StatusCode() != 0 {
		statusCode := fmt.Sprintf("%d", options.StatusCode())
		mw.ResponseCounts[statusCode]++
		mw.TotalResponseCounts[statusCode]++
		mw.TotalResponseTime = mw.TotalResponseTime.Add(responseTime)
		mw.TotalResponseSize += int64(options.Size())
	}
}

// MeasureSince method for execution time recording
func (mw *Stats) MeasureSince(key string, start time.Time) {
	mw.MeasureSinceWithLabels(key, start, nil)
}

// MeasureSinceWithLabels method for execution time recording with custom labels
func (mw *Stats) MeasureSinceWithLabels(key string, start time.Time, labels []Label) {
	labels = append(labels, Label{"host", mw.Hostname})
	elapsed := time.Since(start)

	mw.mu.Lock()
	defer mw.mu.Unlock()

	mw.MetricsCounts[key]++
	mw.MetricsTimers[key] = mw.MetricsTimers[key].Add(elapsed)
}

// Data serializable structure
type Data struct {
	Pid                    int                `json:"pid"`
	Hostname               string             `json:"hostname"`
	UpTime                 string             `json:"uptime"`
	UpTimeSec              float64            `json:"uptime_sec"`
	Time                   string             `json:"time"`
	TimeUnix               int64              `json:"unixtime"`
	StatusCodeCount        map[string]int     `json:"status_code_count"`
	TotalStatusCodeCount   map[string]int     `json:"total_status_code_count"`
	Count                  int                `json:"count"`
	TotalCount             int                `json:"total_count"`
	TotalResponseTime      string             `json:"total_response_time"`
	TotalResponseTimeSec   float64            `json:"total_response_time_sec"`
	TotalResponseSize      int64              `json:"total_response_size"`
	AverageResponseSize    int64              `json:"average_response_size"`
	AverageResponseTime    string             `json:"average_response_time"`
	AverageResponseTimeSec float64            `json:"average_response_time_sec"`
	TotalMetricsCounts     map[string]int     `json:"total_metrics_counts"`
	AverageMetricsTimers   map[string]float64 `json:"average_metrics_timers"`
}

// Data returns the data serializable structure
func (mw *Stats) Data() *Data {
	mw.mu.RLock()

	responseCounts := make(map[string]int, len(mw.ResponseCounts))
	totalResponseCounts := make(map[string]int, len(mw.TotalResponseCounts))
	totalMetricsCounts := make(map[string]int, len(mw.MetricsCounts))
	metricsCounts := make(map[string]float64, len(mw.MetricsCounts))

	now := time.Now()

	uptime := now.Sub(mw.Uptime)

	count := 0
	for code, current := range mw.ResponseCounts {
		responseCounts[code] = current
		count += current
	}

	totalCount := 0
	for code, count := range mw.TotalResponseCounts {
		totalResponseCounts[code] = count
		totalCount += count
	}

	totalResponseTime := mw.TotalResponseTime.Sub(time.Time{})
	totalResponseSize := mw.TotalResponseSize

	averageResponseTime := time.Duration(0)
	averageResponseSize := int64(0)
	if totalCount > 0 {
		avgNs := int64(totalResponseTime) / int64(totalCount)
		averageResponseTime = time.Duration(avgNs)
		averageResponseSize = int64(totalResponseSize) / int64(totalCount)
	}

	for key, count := range mw.MetricsCounts {
		totalMetric := mw.MetricsTimers[key].Sub(time.Time{})
		avgNs := int64(totalMetric) / int64(count)
		metricsCounts[key] = time.Duration(avgNs).Seconds()
		totalMetricsCounts[key] = count
	}

	mw.mu.RUnlock()

	r := &Data{
		Pid:                    mw.Pid,
		UpTime:                 uptime.String(),
		UpTimeSec:              uptime.Seconds(),
		Time:                   now.String(),
		TimeUnix:               now.Unix(),
		StatusCodeCount:        responseCounts,
		TotalStatusCodeCount:   totalResponseCounts,
		Count:                  count,
		TotalCount:             totalCount,
		TotalResponseTime:      totalResponseTime.String(),
		TotalResponseSize:      totalResponseSize,
		TotalResponseTimeSec:   totalResponseTime.Seconds(),
		TotalMetricsCounts:     totalMetricsCounts,
		AverageResponseSize:    averageResponseSize,
		AverageResponseTime:    averageResponseTime.String(),
		AverageResponseTimeSec: averageResponseTime.Seconds(),
		AverageMetricsTimers:   metricsCounts,
	}

	return r
}
