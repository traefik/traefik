// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// Package retry implements flexible retry loops, including support for
// channel cancellation, mocked time, and composable retry strategies
// including exponential backoff with jitter.
//
// The basic usage is as follows:
//
//	for a := retry.Start(someStrategy, nil); a.Next(); {
//		try()
//	}
//
// See examples for details of suggested usage.
package retry // import "gopkg.in/retry.v1"

import (
	"time"
)

// Strategy is implemented by types that represent a retry strategy.
//
// Note: You probably won't need to implement a new strategy - the existing types
// and functions are intended to be sufficient for most purposes.
type Strategy interface {
	// NewTimer is called when the strategy is started - it is
	// called with the time that the strategy is started and returns
	// an object that is used to find out how long to sleep before
	// each retry attempt.
	NewTimer(now time.Time) Timer
}

// Timer represents a source of timing events for a retry strategy.
type Timer interface {
	// NextSleep is called with the time that Next or More has been
	// called and returns the length of time to sleep before the
	// next retry. If no more attempts should be made it should
	// return false, and the returned duration will be ignored.
	//
	// Note that NextSleep is called once after each iteration has
	// completed, assuming the retry loop is continuing.
	NextSleep(now time.Time) (time.Duration, bool)
}

// Attempt represents a running retry attempt.
type Attempt struct {
	clock Clock
	stop  <-chan struct{}
	timer Timer

	// next holds when the next attempt should start.
	// It is valid only when known is true.
	next time.Time

	// count holds the iteration count.
	count int

	// known holds whether next and running are known.
	known bool

	// running holds whether the attempt is still going.
	running bool

	// stopped holds whether the attempt has been stopped.
	stopped bool
}

// Start begins a new sequence of attempts for the given strategy using
// the given Clock implementation for time keeping. If clk is
// nil, the time package will be used to keep time.
func Start(strategy Strategy, clk Clock) *Attempt {
	return StartWithCancel(strategy, clk, nil)
}

// StartWithCancel is like Start except that if a value
// is received on stop while waiting, the attempt will be aborted.
func StartWithCancel(strategy Strategy, clk Clock, stop <-chan struct{}) *Attempt {
	if clk == nil {
		clk = wallClock{}
	}
	now := clk.Now()
	return &Attempt{
		clock:   clk,
		stop:    stop,
		timer:   strategy.NewTimer(now),
		known:   true,
		running: true,
		next:    now,
	}
}

// Next reports whether another attempt should be made, waiting as
// necessary until it's time for the attempt. It always returns true the
// first time it is called unless a value is received on the stop
// channel - we are guaranteed to make at least one attempt unless
// stopped.
func (a *Attempt) Next() bool {
	if !a.More() {
		return false
	}
	sleep := a.next.Sub(a.clock.Now())
	if sleep <= 0 {
		// We're not going to sleep for any length of time,
		// so guarantee that we respect the stop channel. This
		// ensures that we make no attempts if Next is called
		// with a value available on the stop channel.
		select {
		case <-a.stop:
			a.stopped = true
			a.running = false
			return false
		default:
			a.known = false
			a.count++
			return true
		}
	}
	select {
	case <-a.clock.After(sleep):
		a.known = false
		a.count++
	case <-a.stop:
		a.running = false
		a.stopped = true
	}
	return a.running
}

// More reports whether there are more retry attempts to be made. It
// does not sleep.
//
// If More returns false, Next will return false. If More returns true,
// Next will return true except when the attempt has been explicitly
// stopped via the stop channel.
func (a *Attempt) More() bool {
	if !a.known {
		now := a.clock.Now()
		sleepDuration, running := a.timer.NextSleep(now)
		a.next, a.running, a.known = now.Add(sleepDuration), running, true
	}
	return a.running
}

// Stopped reports whether the attempt has terminated because
// a value was received on the stop channel.
func (a *Attempt) Stopped() bool {
	return a.stopped
}

// Count returns the current attempt count number, starting at 1.
// It returns 0 if called before Next is called.
// When the loop has terminated, it holds the total number
// of retries made.
func (a *Attempt) Count() int {
	return a.count
}
