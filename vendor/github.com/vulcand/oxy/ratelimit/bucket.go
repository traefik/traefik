package ratelimit

import (
	"fmt"
	"time"

	"github.com/mailgun/timetools"
)

const UndefinedDelay = -1

// rate defines token bucket parameters.
type rate struct {
	period  time.Duration
	average int64
	burst   int64
}

func (r *rate) String() string {
	return fmt.Sprintf("rate(%v/%v, burst=%v)", r.average, r.period, r.burst)
}

// Implements token bucket algorithm (http://en.wikipedia.org/wiki/Token_bucket)
type tokenBucket struct {
	// The time period controlled by the bucket in nanoseconds.
	period time.Duration
	// The number of nanoseconds that takes to add one more token to the total
	// number of available tokens. It effectively caches the value that could
	// have been otherwise deduced from refillRate.
	timePerToken time.Duration
	// The maximum number of tokens that can be accumulate in the bucket.
	burst int64
	// The number of tokens available for consumption at the moment. It can
	// nether be larger then capacity.
	availableTokens int64
	// Interface that gives current time (so tests can override)
	clock timetools.TimeProvider
	// Tells when tokensAvailable was updated the last time.
	lastRefresh time.Time
	// The number of tokens consumed the last time.
	lastConsumed int64
}

// newTokenBucket crates a `tokenBucket` instance for the specified `Rate`.
func newTokenBucket(rate *rate, clock timetools.TimeProvider) *tokenBucket {
	return &tokenBucket{
		period:          rate.period,
		timePerToken:    time.Duration(int64(rate.period) / rate.average),
		burst:           rate.burst,
		clock:           clock,
		lastRefresh:     clock.UtcNow(),
		availableTokens: rate.burst,
	}
}

// consume makes an attempt to consume the specified number of tokens from the
// bucket. If there are enough tokens available then `0, nil` is returned; if
// tokens to consume is larger than the burst size, then an error is returned
// and the delay is not defined; otherwise returned a none zero delay that tells
// how much time the caller needs to wait until the desired number of tokens
// will become available for consumption.
func (tb *tokenBucket) consume(tokens int64) (time.Duration, error) {
	tb.updateAvailableTokens()
	tb.lastConsumed = 0
	if tokens > tb.burst {
		return UndefinedDelay, fmt.Errorf("Requested tokens larger than max tokens")
	}
	if tb.availableTokens < tokens {
		return tb.timeTillAvailable(tokens), nil
	}
	tb.availableTokens -= tokens
	tb.lastConsumed = tokens
	return 0, nil
}

// rollback reverts effect of the most recent consumption. If the most recent
// `consume` resulted in an error or a burst overflow, and therefore did not
// modify the number of available tokens, then `rollback` won't do that either.
// It is safe to call this method multiple times, for the second and all
// following calls have no effect.
func (tb *tokenBucket) rollback() {
	tb.availableTokens += tb.lastConsumed
	tb.lastConsumed = 0
}

// Update modifies `average` and `burst` fields of the token bucket according
// to the provided `Rate`
func (tb *tokenBucket) update(rate *rate) error {
	if rate.period != tb.period {
		return fmt.Errorf("Period mismatch: %v != %v", tb.period, rate.period)
	}
	tb.timePerToken = time.Duration(int64(tb.period) / rate.average)
	tb.burst = rate.burst
	if tb.availableTokens > rate.burst {
		tb.availableTokens = rate.burst
	}
	return nil
}

// timeTillAvailable returns the number of nanoseconds that we need to
// wait until the specified number of tokens becomes available for consumption.
func (tb *tokenBucket) timeTillAvailable(tokens int64) time.Duration {
	missingTokens := tokens - tb.availableTokens
	return time.Duration(missingTokens) * tb.timePerToken
}

// updateAvailableTokens updates the number of tokens available for consumption.
// It is calculated based on the refill rate, the time passed since last refresh,
// and is limited by the bucket capacity.
func (tb *tokenBucket) updateAvailableTokens() {
	now := tb.clock.UtcNow()
	timePassed := now.Sub(tb.lastRefresh)

	tokens := tb.availableTokens + int64(timePassed/tb.timePerToken)
	// If we haven't added any tokens that means that not enough time has passed,
	// in this case do not adjust last refill checkpoint, otherwise it will be
	// always moving in time in case of frequent requests that exceed the rate
	if tokens != tb.availableTokens {
		tb.lastRefresh = now
		tb.availableTokens = tokens
	}
	if tb.availableTokens > tb.burst {
		tb.availableTokens = tb.burst
	}
}
