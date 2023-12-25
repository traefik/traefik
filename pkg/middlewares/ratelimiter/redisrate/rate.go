package redisrate

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
	rate := limit.Rate / 1000000
	t := time.Now().UnixMicro()
	params := []interface{}{float64(rate), limit.Burst, l.ttl, t, float64(n), l.maxDelay.Microseconds()}
	v, err := allowTokenBucket.Run(ctx, l.rdb, []string{redisPrefix + key}, params...).Result()
	if err != nil {
		return nil, err
	}
	values := v.([]interface{})

	ok, err := strconv.ParseBool(values[0].(string))
	if err != nil {
		return nil, err
	}
	delay, err := strconv.ParseFloat(values[1].(string), 64)
	if err != nil {
		return nil, err
	}

	res := &Result{
		Ok:    ok,
		Delay: dur(delay),
	}
	return res, nil
}

func dur(f float64) time.Duration {
	if f == -1 {
		return -1
	}
	return time.Duration(f * float64(time.Microsecond))
}

type Result struct {
	// If meet bursty traffic or not.
	Ok bool
	// Delay for handling request.
	Delay time.Duration
}
