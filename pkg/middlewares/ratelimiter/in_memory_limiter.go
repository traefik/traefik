package ratelimiter

import (
	"context"
	"time"

	"github.com/mailgun/ttlmap"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type InMemoryRateLimiter struct {
	rate  rate.Limit // reqs/s
	burst int64
	// maxDelay is the maximum duration we're willing to wait for a bucket reservation to become effective, in nanoseconds.
	// For now it is somewhat arbitrarily set to 1/(2*rate).
	maxDelay time.Duration
	// Each rate limiter for a given source is stored in the buckets ttlmap.
	// To keep this ttlmap constrained in size,
	// each ratelimiter is "garbage collected" when it is considered expired.
	// It is considered expired after it hasn't been used for ttl seconds.
	ttl     int
	buckets *ttlmap.TtlMap // actual buckets, keyed by source.

	logger *zerolog.Logger
}

func NewInMemoryRateLimiter(
	rate rate.Limit,
	burst int64,
	maxDelay time.Duration,
	ttl int,
	logger *zerolog.Logger,
) (Limiter, error) {
	buckets, err := ttlmap.NewConcurrent(maxSources)
	if err != nil {
		return nil, err
	}
	return &InMemoryRateLimiter{
		rate:     rate,
		burst:    burst,
		maxDelay: maxDelay,
		ttl:      ttl,
		logger:   logger,

		buckets: buckets,
	}, nil
}

func (b *InMemoryRateLimiter) Allow(ctx context.Context, source string) (Result, error) {
	// Get bucket which contains limiter information.
	var bucket *rate.Limiter
	if rlSource, exists := b.buckets.Get(source); exists {
		bucket = rlSource.(*rate.Limiter)
	} else {
		bucket = rate.NewLimiter(b.rate, int(b.burst))
	}

	// We Set even in the case where the source already exists,
	// because we want to update the expiryTime everytime we get the source,
	// as the expiryTime is supposed to reflect the activity (or lack thereof) on that source.
	if err := b.buckets.Set(source, bucket, b.ttl); err != nil {
		return Result{}, err
	}

	res := bucket.Reserve()
	if !res.OK() {
		return Result{
			Ok: false,
		}, nil
	}
	delay := res.Delay()
	if delay > b.maxDelay {
		res.Cancel()
		return Result{
			Ok:    false,
			Delay: delay,
		}, nil
	}

	select {
	case <-ctx.Done():
		return Result{Ok: false}, nil
	case <-time.After(delay):
	}

	return Result{
		Ok:        res.OK(),
		Remaining: bucket.Tokens(),
		Delay:     delay,
	}, nil
}
