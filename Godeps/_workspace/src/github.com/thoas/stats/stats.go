package stats

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Stats struct {
	mu                  sync.RWMutex
	Uptime              time.Time
	Pid                 int
	ResponseCounts      map[string]int
	TotalResponseCounts map[string]int
	TotalResponseTime   time.Time
}

func New() *Stats {
	stats := &Stats{
		Uptime:              time.Now(),
		Pid:                 os.Getpid(),
		ResponseCounts:      map[string]int{},
		TotalResponseCounts: map[string]int{},
		TotalResponseTime:   time.Time{},
	}

	go func() {
		for {
			stats.ResetResponseCounts()

			time.Sleep(time.Second * 1)
		}
	}()

	return stats
}

func (mw *Stats) ResetResponseCounts() {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.ResponseCounts = map[string]int{}
}

// MiddlewareFunc makes Stats implement the Middleware interface.
func (mw *Stats) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beginning, recorder := mw.Begin(w)

		h.ServeHTTP(recorder, r)

		mw.End(beginning, recorder)
	})
}

// Negroni compatible interface
func (mw *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	beginning, recorder := mw.Begin(w)

	next(recorder, r)

	mw.End(beginning, recorder)
}

func (mw *Stats) Begin(w http.ResponseWriter) (time.Time, Recorder) {
	start := time.Now()

	writer := &RecorderResponseWriter{w, 200, 0}

	return start, writer
}

func (mw *Stats) EndWithStatus(start time.Time, status int) {
	end := time.Now()

	responseTime := end.Sub(start)

	mw.mu.Lock()

	defer mw.mu.Unlock()

	statusCode := fmt.Sprintf("%d", status)

	mw.ResponseCounts[statusCode]++
	mw.TotalResponseCounts[statusCode]++
	mw.TotalResponseTime = mw.TotalResponseTime.Add(responseTime)
}

func (mw *Stats) End(start time.Time, recorder Recorder) {
	mw.EndWithStatus(start, recorder.Status())
}

type data struct {
	Pid                    int            `json:"pid"`
	UpTime                 string         `json:"uptime"`
	UpTimeSec              float64        `json:"uptime_sec"`
	Time                   string         `json:"time"`
	TimeUnix               int64          `json:"unixtime"`
	StatusCodeCount        map[string]int `json:"status_code_count"`
	TotalStatusCodeCount   map[string]int `json:"total_status_code_count"`
	Count                  int            `json:"count"`
	TotalCount             int            `json:"total_count"`
	TotalResponseTime      string         `json:"total_response_time"`
	TotalResponseTimeSec   float64        `json:"total_response_time_sec"`
	AverageResponseTime    string         `json:"average_response_time"`
	AverageResponseTimeSec float64        `json:"average_response_time_sec"`
}

func (mw *Stats) Data() *data {

	mw.mu.RLock()

	now := time.Now()

	uptime := now.Sub(mw.Uptime)

	count := 0
	for _, current := range mw.ResponseCounts {
		count += current
	}

	totalCount := 0
	for _, count := range mw.TotalResponseCounts {
		totalCount += count
	}

	totalResponseTime := mw.TotalResponseTime.Sub(time.Time{})

	averageResponseTime := time.Duration(0)
	if totalCount > 0 {
		avgNs := int64(totalResponseTime) / int64(totalCount)
		averageResponseTime = time.Duration(avgNs)
	}

	r := &data{
		Pid:                    mw.Pid,
		UpTime:                 uptime.String(),
		UpTimeSec:              uptime.Seconds(),
		Time:                   now.String(),
		TimeUnix:               now.Unix(),
		StatusCodeCount:        mw.ResponseCounts,
		TotalStatusCodeCount:   mw.TotalResponseCounts,
		Count:                  count,
		TotalCount:             totalCount,
		TotalResponseTime:      totalResponseTime.String(),
		TotalResponseTimeSec:   totalResponseTime.Seconds(),
		AverageResponseTime:    averageResponseTime.String(),
		AverageResponseTimeSec: averageResponseTime.Seconds(),
	}

	mw.mu.RUnlock()

	return r
}
