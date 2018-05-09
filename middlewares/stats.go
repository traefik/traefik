package middlewares

import (
	"bufio"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	_ Stateful = &responseRecorder{}
)

// StatsRecorder is an optional middleware that records more details statistics
// about requests and how they are processed. This currently consists of recent
// requests that have caused errors (4xx and 5xx status codes), making it easy
// to pinpoint problems.
type StatsRecorder struct {
	mutex           sync.RWMutex
	numRecentErrors int
	recentErrors    []*statsError
}

// NewStatsRecorder returns a new StatsRecorder
func NewStatsRecorder(numRecentErrors int) *StatsRecorder {
	return &StatsRecorder{
		numRecentErrors: numRecentErrors,
	}
}

// Stats includes all of the stats gathered by the recorder.
type Stats struct {
	RecentErrors []*statsError `json:"recent_errors"`
}

// statsError represents an error that has occurred during request processing.
type statsError struct {
	StatusCode int       `json:"status_code"`
	Status     string    `json:"status"`
	Method     string    `json:"method"`
	Host       string    `json:"host"`
	Path       string    `json:"path"`
	Time       time.Time `json:"time"`
}

// responseRecorder captures information from the response and preserves it for
// later analysis.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code for later retrieval.
func (r *responseRecorder) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.statusCode = status
}

// Hijack hijacks the connection
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone
// away.
func (r *responseRecorder) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush sends any buffered data to the client.
func (r *responseRecorder) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

// ServeHTTP silently extracts information from the request and response as it
// is processed. If the response is 4xx or 5xx, add it to the list of 10 most
// recent errors.
func (s *StatsRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	recorder := &responseRecorder{w, http.StatusOK}
	next(recorder, r)
	if recorder.statusCode >= http.StatusBadRequest {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.recentErrors = append([]*statsError{
			{
				StatusCode: recorder.statusCode,
				Status:     http.StatusText(recorder.statusCode),
				Method:     r.Method,
				Host:       r.Host,
				Path:       r.URL.Path,
				Time:       time.Now(),
			},
		}, s.recentErrors...)
		// Limit the size of the list to numRecentErrors
		if len(s.recentErrors) > s.numRecentErrors {
			s.recentErrors = s.recentErrors[:s.numRecentErrors]
		}
	}
}

// Data returns a copy of the statistics that have been gathered.
func (s *StatsRecorder) Data() *Stats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// We can't return the slice directly or a race condition might develop
	recentErrors := make([]*statsError, len(s.recentErrors))
	copy(recentErrors, s.recentErrors)

	return &Stats{
		RecentErrors: recentErrors,
	}
}
