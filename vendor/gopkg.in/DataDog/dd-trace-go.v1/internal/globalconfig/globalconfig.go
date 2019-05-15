// Package globalconfig stores configuration which applies globally to both the tracer
// and integrations.
package globalconfig

import "sync"

var cfg = &config{}

type config struct {
	mu            sync.RWMutex
	analyticsRate float64
}

// AnalyticsRate returns the sampling rate at which events should be marked. It uses
// synchronizing mechanisms, meaning that for optimal performance it's best to read it
// once and store it.
func AnalyticsRate() float64 {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	return cfg.analyticsRate
}

// SetAnalyticsRate sets the given event sampling rate globally.
func SetAnalyticsRate(rate float64) {
	cfg.mu.Lock()
	cfg.analyticsRate = rate
	cfg.mu.Unlock()
}
