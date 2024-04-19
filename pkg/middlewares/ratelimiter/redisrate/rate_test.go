package redisrate_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func rateLimiter() *redis_rate.Limiter {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{"server0": "localhost:6379"},
	})
	if err := ring.FlushDB(context.TODO()).Err(); err != nil {
		panic(err)
	}
	return redis_rate.NewLimiter(ring)
}

func TestAllow(t *testing.T) {
	ctx := context.Background()

	l := rateLimiter()

	limit := redis_rate.PerSecond(10)
	require.Equal(t, "10 req/s (burst 10)", limit.String())
	require.False(t, limit.IsZero())

	res, err := l.Allow(ctx, "test_id", limit)
	require.NoError(t, err)
	require.Equal(t, 1, res.Allowed)
	require.Equal(t, 9, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 100*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	err = l.Reset(ctx, "test_id")
	require.NoError(t, err)
	res, err = l.Allow(ctx, "test_id", limit)
	require.NoError(t, err)
	require.Equal(t, 1, res.Allowed)
	require.Equal(t, 9, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 100*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowN(ctx, "test_id", limit, 2)
	require.NoError(t, err)
	require.Equal(t, 2, res.Allowed)
	require.Equal(t, 7, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 300*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowN(ctx, "test_id", limit, 7)
	require.NoError(t, err)
	require.Equal(t, 7, res.Allowed)
	require.Equal(t, 0, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 999*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowN(ctx, "test_id", limit, 1000)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 0, res.Remaining)
	require.InDelta(t, 99*time.Second, res.RetryAfter, float64(time.Second))
	require.InDelta(t, 999*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))
}

func TestAllowN_IncrementZero(t *testing.T) {
	ctx := context.Background()
	l := rateLimiter()
	limit := redis_rate.PerSecond(10)

	// Check for a row that's not there
	res, err := l.AllowN(ctx, "test_id", limit, 0)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 10, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.Equal(t, time.Duration(0), res.ResetAfter)

	// Now increment it
	res, err = l.Allow(ctx, "test_id", limit)
	require.NoError(t, err)
	require.Equal(t, 1, res.Allowed)
	require.Equal(t, 9, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 100*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	// Peek again
	res, err = l.AllowN(ctx, "test_id", limit, 0)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 9, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 100*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))
}

func TestRetryAfter(t *testing.T) {
	limit := redis_rate.Limit{
		Rate:   1,
		Period: time.Millisecond,
		Burst:  1,
	}

	ctx := context.Background()
	l := rateLimiter()

	for i := 0; i < 1000; i++ {
		res, err := l.Allow(ctx, "test_id", limit)
		require.NoError(t, err)

		if res.Allowed > 0 {
			continue
		}

		require.LessOrEqual(t, int64(res.RetryAfter), int64(time.Millisecond))
	}
}

func TestAllowAtMost(t *testing.T) {
	ctx := context.Background()

	l := rateLimiter()
	limit := redis_rate.PerSecond(10)

	res, err := l.Allow(ctx, "test_id", limit)
	require.NoError(t, err)
	require.Equal(t, 1, res.Allowed)
	require.Equal(t, 9, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 100*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowAtMost(ctx, "test_id", limit, 2)
	require.NoError(t, err)
	require.Equal(t, 2, res.Allowed)
	require.Equal(t, 7, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 300*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowN(ctx, "test_id", limit, 0)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 7, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 300*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowAtMost(ctx, "test_id", limit, 10)
	require.NoError(t, err)
	require.Equal(t, 7, res.Allowed)
	require.Equal(t, 0, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 999*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowN(ctx, "test_id", limit, 0)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 0, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 999*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowAtMost(ctx, "test_id", limit, 1000)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 0, res.Remaining)
	require.InDelta(t, 99*time.Millisecond, res.RetryAfter, float64(10*time.Millisecond))
	require.InDelta(t, 999*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	res, err = l.AllowN(ctx, "test_id", limit, 1000)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 0, res.Remaining)
	require.InDelta(t, 99*time.Second, res.RetryAfter, float64(time.Second))
	require.InDelta(t, 999*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))
}

func TestAllowAtMost_IncrementZero(t *testing.T) {
	ctx := context.Background()
	l := rateLimiter()
	limit := redis_rate.PerSecond(10)

	// Check for a row that isn't there
	res, err := l.AllowAtMost(ctx, "test_id", limit, 0)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 10, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.Equal(t, time.Duration(0), res.ResetAfter)

	// Now increment it
	res, err = l.Allow(ctx, "test_id", limit)
	require.NoError(t, err)
	require.Equal(t, 1, res.Allowed)
	require.Equal(t, 9, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 100*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))

	// Peek again
	res, err = l.AllowAtMost(ctx, "test_id", limit, 0)
	require.NoError(t, err)
	require.Equal(t, 0, res.Allowed)
	require.Equal(t, 9, res.Remaining)
	require.Equal(t, time.Duration(-1), res.RetryAfter)
	require.InDelta(t, 100*time.Millisecond, res.ResetAfter, float64(10*time.Millisecond))
}

func BenchmarkAllow(b *testing.B) {
	ctx := context.Background()
	l := rateLimiter()
	limit := redis_rate.PerSecond(1e6)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := l.Allow(ctx, "foo", limit)
			if err != nil {
				b.Fatal(err)
			}
			if res.Allowed == 0 {
				panic("not reached")
			}
		}
	})
}

func BenchmarkAllowAtMost(b *testing.B) {
	ctx := context.Background()
	l := rateLimiter()
	limit := redis_rate.PerSecond(1e6)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := l.AllowAtMost(ctx, "foo", limit, 1)
			if err != nil {
				b.Fatal(err)
			}
			if res.Allowed == 0 {
				panic("not reached")
			}
		}
	})
}
