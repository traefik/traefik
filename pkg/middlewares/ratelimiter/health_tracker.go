package ratelimiter

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// healthTracker tracks the health status of the rate limiter
type healthTracker struct {
	mu               sync.RWMutex
	isShutdown       bool
	shutdownUntil    time.Time
	failureCount     int
	lastFailureReset time.Time
	backoffTimeout   time.Duration
	backoffDuration  time.Duration
	backoffThreshold int
	logger           *zerolog.Logger
}

// newHealthTracker creates a new health tracker with the given configuration
func newHealthTracker(backoffTimeout, backoffDuration time.Duration, backoffThreshold int, logger *zerolog.Logger) *healthTracker {
	return &healthTracker{
		backoffTimeout:   backoffTimeout,
		backoffDuration:  backoffDuration,
		backoffThreshold: backoffThreshold,
		logger:           logger,
	}
}

// recordFailure records a failure and checks if the limiter should be shut down
func (ht *healthTracker) recordFailure() bool {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	now := time.Now()

	// Reset failure count if the backoff duration has passed
	if now.Sub(ht.lastFailureReset) > ht.backoffDuration {
		ht.failureCount = 0
		ht.lastFailureReset = now
	}

	ht.failureCount++

	// Check if we should shut down the limiter
	// Only shutdown if threshold is non-negative and we've reached it
	if ht.backoffThreshold >= 0 && ht.failureCount >= ht.backoffThreshold {
		ht.isShutdown = true
		ht.shutdownUntil = now.Add(ht.backoffTimeout)
		ht.logger.Warn().
			Int("failureCount", ht.failureCount).
			Dur("shutdownUntil", ht.backoffTimeout).
			Msg("Rate limiter shut down due to repeated failures")
		return true
	}

	return false
}

// isShutdownNow checks if the limiter is currently shut down
func (ht *healthTracker) isShutdownNow() bool {
	// Fast path: lockless read for performance in the hot path
	// This may occasionally read a stale value during state transitions,
	// but this is acceptable for rate limiting where perfect precision isn't critical
	if !ht.isShutdown {
		return false
	}

	// Check if shutdown period has expired
	if ht.isShutdown && time.Now().After(ht.shutdownUntil) {
		ht.mu.Lock()
		defer ht.mu.Unlock()
		// Double-check after acquiring write lock
		if ht.isShutdown && time.Now().After(ht.shutdownUntil) {
			ht.isShutdown = false
			ht.failureCount = 0
			ht.lastFailureReset = time.Now()
			ht.logger.Info().Msg("Rate limiter recovered from shutdown")
		}
		return false
	}

	return ht.isShutdown
}

// getStatus returns the current status of the health tracker for testing purposes
func (ht *healthTracker) getStatus() (isShutdown bool, failureCount int, shutdownUntil time.Time) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.isShutdown, ht.failureCount, ht.shutdownUntil
}

// getThreshold returns the backoff threshold for testing purposes
func (ht *healthTracker) getThreshold() int {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.backoffThreshold
}
