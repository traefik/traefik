package ratelimiter

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthTracker(t *testing.T) {
	logger := zerolog.Nop()
	backoffTimeout := 30 * time.Second
	backoffDuration := 10 * time.Second
	backoffThreshold := 5

	tracker := newHealthTracker(backoffTimeout, backoffDuration, backoffThreshold, &logger)

	assert.Equal(t, backoffTimeout, tracker.backoffTimeout)
	assert.Equal(t, backoffDuration, tracker.backoffDuration)
	assert.Equal(t, backoffThreshold, tracker.backoffThreshold)
	assert.Equal(t, &logger, tracker.logger)
	assert.False(t, tracker.isShutdown)
	assert.Equal(t, 0, tracker.failureCount)
}

func TestRecordFailure_UnderThreshold(t *testing.T) {
	logger := zerolog.Nop()
	tracker := newHealthTracker(30*time.Second, 10*time.Second, 3, &logger)

	// Record 2 failures (under threshold)
	shouldShutdown := tracker.recordFailure()
	assert.False(t, shouldShutdown)
	assert.False(t, tracker.isShutdown)
	assert.Equal(t, 1, tracker.failureCount)

	shouldShutdown = tracker.recordFailure()
	assert.False(t, shouldShutdown)
	assert.False(t, tracker.isShutdown)
	assert.Equal(t, 2, tracker.failureCount)
}

func TestRecordFailure_AtThreshold(t *testing.T) {
	logger := zerolog.Nop()
	tracker := newHealthTracker(30*time.Second, 10*time.Second, 3, &logger)

	// Record 3 failures (at threshold)
	shouldShutdown := tracker.recordFailure()
	assert.False(t, shouldShutdown)
	assert.Equal(t, 1, tracker.failureCount)

	shouldShutdown = tracker.recordFailure()
	assert.False(t, shouldShutdown)
	assert.Equal(t, 2, tracker.failureCount)

	shouldShutdown = tracker.recordFailure()
	assert.True(t, shouldShutdown)
	assert.True(t, tracker.isShutdown)
	assert.Equal(t, 3, tracker.failureCount)
}

func TestRecordFailure_OverThreshold(t *testing.T) {
	logger := zerolog.Nop()
	tracker := newHealthTracker(30*time.Second, 10*time.Second, 2, &logger)

	// Record 3 failures (over threshold)
	shouldShutdown := tracker.recordFailure()
	assert.False(t, shouldShutdown)

	shouldShutdown = tracker.recordFailure()
	assert.True(t, shouldShutdown)
	assert.True(t, tracker.isShutdown)

	// Additional failures after shutdown should still return true
	shouldShutdown = tracker.recordFailure()
	assert.True(t, shouldShutdown)
	assert.True(t, tracker.isShutdown)
}

func TestRecordFailure_ResetAfterPeriod(t *testing.T) {
	logger := zerolog.Nop()
	backoffDuration := 100 * time.Millisecond
	tracker := newHealthTracker(30*time.Second, backoffDuration, 2, &logger)

	// Record 1 failure
	shouldShutdown := tracker.recordFailure()
	assert.False(t, shouldShutdown)
	assert.Equal(t, 1, tracker.failureCount)

	// Wait for the backoff duration to expire
	time.Sleep(backoffDuration + 10*time.Millisecond)

	// Record another failure - should reset counter
	shouldShutdown = tracker.recordFailure()
	assert.False(t, shouldShutdown)
	assert.Equal(t, 1, tracker.failureCount) // Reset to 1, not 2
}

func TestIsShutdownNow_NotShutdown(t *testing.T) {
	logger := zerolog.Nop()
	tracker := newHealthTracker(30*time.Second, 10*time.Second, 2, &logger)

	assert.False(t, tracker.isShutdownNow())
}

func TestIsShutdownNow_CurrentlyShutdown(t *testing.T) {
	logger := zerolog.Nop()
	backoffTimeout := 100 * time.Millisecond
	tracker := newHealthTracker(backoffTimeout, 10*time.Second, 1, &logger)

	// Trigger shutdown
	shouldShutdown := tracker.recordFailure()
	require.True(t, shouldShutdown)
	require.True(t, tracker.isShutdown)

	// Should still be shutdown
	assert.True(t, tracker.isShutdownNow())
}

func TestIsShutdownNow_RecoveryAfterTimeout(t *testing.T) {
	logger := zerolog.Nop()
	backoffTimeout := 50 * time.Millisecond
	tracker := newHealthTracker(backoffTimeout, 10*time.Second, 1, &logger)

	// Trigger shutdown
	shouldShutdown := tracker.recordFailure()
	require.True(t, shouldShutdown)
	require.True(t, tracker.isShutdown)

	// Wait for backoff timeout to expire
	time.Sleep(backoffTimeout + 10*time.Millisecond)

	// Should have recovered
	assert.False(t, tracker.isShutdownNow())

	// Check internal state
	isShutdown, failureCount, _ := tracker.getStatus()
	assert.False(t, isShutdown)
	assert.Equal(t, 0, failureCount)
}

func TestConcurrentAccess(t *testing.T) {
	logger := zerolog.Nop()
	tracker := newHealthTracker(30*time.Second, 10*time.Second, 10, &logger)

	// Test concurrent recordFailure calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			tracker.recordFailure()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should be shutdown after 10 failures
	assert.True(t, tracker.isShutdownNow())
}

func TestGetStatus(t *testing.T) {
	logger := zerolog.Nop()
	backoffTimeout := 30 * time.Second
	tracker := newHealthTracker(backoffTimeout, 10*time.Second, 2, &logger)

	// Initial state
	isShutdown, failureCount, shutdownUntil := tracker.getStatus()
	assert.False(t, isShutdown)
	assert.Equal(t, 0, failureCount)
	assert.True(t, shutdownUntil.IsZero())

	// Record one failure
	tracker.recordFailure()
	isShutdown, failureCount, shutdownUntil = tracker.getStatus()
	assert.False(t, isShutdown)
	assert.Equal(t, 1, failureCount)
	assert.True(t, shutdownUntil.IsZero())

	// Record second failure to trigger shutdown
	tracker.recordFailure()
	isShutdown, failureCount, shutdownUntil = tracker.getStatus()
	assert.True(t, isShutdown)
	assert.Equal(t, 2, failureCount)
	assert.False(t, shutdownUntil.IsZero())
	assert.True(t, shutdownUntil.After(time.Now()))
}

func TestEdgeCase_ZeroThreshold(t *testing.T) {
	logger := zerolog.Nop()
	tracker := newHealthTracker(30*time.Second, 10*time.Second, 0, &logger)

	// With threshold 0, first failure should trigger shutdown
	shouldShutdown := tracker.recordFailure()
	assert.True(t, shouldShutdown)
	assert.True(t, tracker.isShutdown)
}

func TestEdgeCase_NegativeThreshold(t *testing.T) {
	logger := zerolog.Nop()
	tracker := newHealthTracker(30*time.Second, 10*time.Second, -1, &logger)

	// With negative threshold, should never shutdown
	shouldShutdown := tracker.recordFailure()
	assert.False(t, shouldShutdown)
	assert.False(t, tracker.isShutdown)
}
