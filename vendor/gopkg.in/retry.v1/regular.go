// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package retry // import "gopkg.in/retry.v1"

import (
	"time"
)

// Regular represents a strategy that repeats at regular intervals.
type Regular struct {
	// Total specifies the total duration of the attempt.
	Total time.Duration

	// Delay specifies the interval between the start of each try
	// in the burst. If an try takes longer than Delay, the
	// next try will happen immediately.
	Delay time.Duration

	// Min holds the minimum number of retries. It overrides Total.
	// To limit the maximum number of retries, use LimitCount.
	Min int
}

// regularTimer holds a running instantiation of the Regular timer.
type regularTimer struct {
	strategy Regular
	count    int
	// start holds when the current try started.
	start time.Time
	end   time.Time
}

// Start is short for Start(r, clk, nil)
func (r Regular) Start(clk Clock) *Attempt {
	return Start(r, clk)
}

// NewTimer implements Strategy.NewTimer.
func (r Regular) NewTimer(now time.Time) Timer {
	return &regularTimer{
		strategy: r,
		start:    now,
		end:      now.Add(r.Total),
	}
}

// NextSleep implements Timer.NextSleep.
func (a *regularTimer) NextSleep(now time.Time) (time.Duration, bool) {
	sleep := a.strategy.Delay - now.Sub(a.start)
	if sleep <= 0 {
		sleep = 0
	}
	a.count++
	// Set the start of the next try.
	a.start = now.Add(sleep)
	if a.count < a.strategy.Min {
		return sleep, true
	}
	// The next try is not before the end - no more attempts.
	if !a.start.Before(a.end) {
		return 0, false
	}
	return sleep, true
}
