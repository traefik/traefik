package redis_rate

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

const redisPrefix = "rate:"

// Inf is the infinite rate limit; it allows all events (even if burst is zero).
const Inf = rate.Limit(math.MaxFloat64)

// InfDuration is the duration returned by Delay when a Reservation is not OK.
const InfDuration = time.Duration(math.MaxInt64)

type rediser interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd

	EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
}

type Limit struct {
	Rate   rate.Limit
	Burst  int64
	Period time.Duration

	tokens float64
	// last is the last time the limiter's tokens field was updated
	last time.Time
	// lastEvent is the latest time of a rate-limited event (past or future)
	lastEvent time.Time
}

func (l Limit) String() string {
	return fmt.Sprintf("%f req/%s (burst %d)", l.Rate, fmtDur(l.Period), l.Burst)
}

func (l Limit) IsZero() bool {
	return l == Limit{}
}

func fmtDur(d time.Duration) string {
	switch d {
	case time.Second:
		return "s"
	case time.Minute:
		return "m"
	case time.Hour:
		return "h"
	}
	return d.String()
}

func PerSecond(rate rate.Limit) Limit {
	return Limit{
		Rate:   rate,
		Period: time.Second,
		Burst:  int64(rate),
	}
}

func PerMinute(rate rate.Limit) Limit {
	return Limit{
		Rate:   rate,
		Period: time.Minute,
		Burst:  int64(rate),
	}
}

func PerHour(rate rate.Limit) Limit {
	return Limit{
		Rate:   rate,
		Period: time.Hour,
		Burst:  int64(rate),
	}
}

// ------------------------------------------------------------------------------

// Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	rdb      rediser
	ttl      int
	maxDelay time.Duration
}

// NewLimiter returns a new Limiter.
func NewLimiter(rdb rediser, ttl int, maxDelay time.Duration) *Limiter {
	return &Limiter{
		rdb:      rdb,
		ttl:      ttl,
		maxDelay: maxDelay,
	}
}

// Allow is a shortcut for AllowN(ctx, key, limit, 1).
func (l Limiter) Allow(ctx context.Context, key string, limit Limit) (*Result, error) {
	return l.AllowTokenBucketN(ctx, key, limit, 1)
}

func (l Limiter) AllowTokenBucketN(ctx context.Context, key string, limit Limit, n int) (*Result, error) {
	rate := limit.Rate / 1000
	t := time.Now()
	params := []interface{}{float64(rate), limit.Burst, l.ttl, t.UnixMilli(), float64(n), l.maxDelay.Milliseconds()}
	v, err := allowTokenBucket.Run(ctx, l.rdb, []string{redisPrefix + key}, params...).Result()
	if err != nil {
		return nil, err
	}
	values := v.([]interface{})

	ok, err := strconv.ParseBool(values[0].(string))
	if err != nil {
		return nil, err
	}
	tokens, err := strconv.ParseFloat(values[1].(string), 64)
	if err != nil {
		return nil, err
	}
	waitDuration, err := strconv.ParseFloat(values[2].(string), 64)
	if err != nil {
		return nil, err
	}
	delay, err := strconv.ParseFloat(values[3].(string), 64)
	if err != nil {
		return nil, err
	}

	res := &Result{
		Ok:           ok,
		Tokens:       tokens,
		WaitDuration: waitDuration,
		Delay:        dur(delay),
	}
	return res, nil

}

func dur(f float64) time.Duration {
	if f == -1 {
		return -1
	}
	return time.Duration(f * float64(time.Millisecond))
}

type Result struct {
	Ok        bool
	timeToAct time.Time
	last      time.Time
	lastEvent time.Time

	Tokens       float64
	WaitDuration float64
	Delay        time.Duration

	// Limit is the limit that was used to obtain this result.
	Limit Limit

	// Allowed is the number of events that may happen at time now.
	// Allowed int

	// Remaining is the maximum number of requests that could be
	// permitted instantaneously for this key given the current
	// state. For example, if a rate limiter allows 10 requests per
	// second and has already received 6 requests for this key this
	// second, Remaining would be 4.
	Remaining float64

	// RetryAfter is the time until the next request will be permitted.
	// It should be -1 unless the rate limit has been exceeded.
	RetryAfter time.Duration

	// ResetAfter is the time until the RateLimiter returns to its
	// initial state for a given key. For example, if a rate limiter
	// manages requests per second and received one request 200ms ago,
	// Reset would return 800ms. You can also think of this as the time
	// until Limit and Remaining will be equal.
	// ResetAfter time.Durationx
}

// func (r *Result) Delay() time.Duration {
// 	if !r.Ok {
// 		return rate.InfDuration
// 	}

// 	now := time.Now()
// 	delay := r.timeToAct.Sub(now)
// 	if delay < 0 {
// 		return 0
// 	}
// 	return delay
// }

// // Cancel is shorthand for CancelAt(time.Now()).
// func (r *Result) Cancel() {
// 	r.CancelAt(time.Now())
// }

// CancelAt indicates that the reservation holder will not perform the reserved action
// and reverses the effects of this Reservation on the rate limit as much as possible,
// considering that other reservations may have already been made.
func (r *Result) CancelAt(t time.Time) {
	if !r.Ok {
		return
	}

	if r.Limit.Rate == Inf || r.Remaining == 0 || r.timeToAct.Before(t) {
		return
	}

	// calculate tokens to restore
	// The duration between lim.lastEvent and r.timeToAct tells us how many tokens were reserved
	// after r was obtained. These tokens should not be restored.
	restoreTokens := float64(r.Remaining) - r.Limit.tokensFromDuration(r.lastEvent.Sub(r.timeToAct))
	if restoreTokens <= 0 {
		return
	}

	// advance time to now
	t, tokens := r.advance(t)
	tokens += restoreTokens

	if burst := float64(r.Limit.Burst); tokens > burst {
		tokens = burst
	}

	// TODO: upadate state
	r.last = t
	r.Remaining = tokens
	if r.timeToAct == r.lastEvent {
		prevEvent := r.timeToAct.Add(r.Limit.durationFromTokens(float64(-r.Remaining)))
		if !prevEvent.Before(t) {
			r.lastEvent = prevEvent
		}
	}

}

// advance calculates and returns an updated state for lim resulting from the passage of time.
// lim is not changed.
// advance requires that lim.mu is held.
func (r *Result) advance(t time.Time) (newT time.Time, newTokens float64) {
	last := r.Limit.last
	//! Just in case avoid negative value.
	if t.Before(last) {
		last = t
	}

	// Calculate the new number of tokens, due to time that passed.
	elapsed := t.Sub(last)
	delta := r.Limit.tokensFromDuration(elapsed)
	tokens := r.Limit.tokens + delta
	if burst := float64(r.Limit.Burst); tokens > burst {
		tokens = burst
	}
	return t, tokens
}

// tokensFromDuration is a unit conversion function from a time duration to the number of tokens
// which could be accumulated during that duration at a rate of limit tokens per second.
func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	if limit.Rate <= 0 {
		return 0
	}
	return d.Seconds() * float64(limit.Rate)
}

// durationFromTokens is a unit conversion function from the number of tokens to the duration
// of time it takes to accumulate them at a rate of limit tokens per second.
func (limit Limit) durationFromTokens(tokens float64) time.Duration {
	if limit.Rate <= 0 {
		return InfDuration
	}
	seconds := tokens / float64(limit.Rate)
	return time.Duration(float64(time.Second) * seconds)
}
