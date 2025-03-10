package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/ttlmap"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type inMemoryRateLimiter struct {
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

func newInMemoryRateLimiter(rate rate.Limit, burst int64, maxDelay time.Duration, ttl int, logger *zerolog.Logger) (*inMemoryRateLimiter, error) {
	buckets, err := ttlmap.NewConcurrent(maxSources)
	if err != nil {
		return nil, fmt.Errorf("creating ttlmap: %w", err)
	}

	return &inMemoryRateLimiter{
		rate:     rate,
		burst:    burst,
		maxDelay: maxDelay,
		ttl:      ttl,
		logger:   logger,
		buckets:  buckets,
	}, nil
}

func (i *inMemoryRateLimiter) Allow(_ context.Context, source string) (*time.Duration, error) {
	// Get bucket which contains limiter information.
	var bucket *rate.Limiter
	if rlSource, exists := i.buckets.Get(source); exists {
		bucket = rlSource.(*rate.Limiter)
	} else {
		bucket = rate.NewLimiter(i.rate, int(i.burst))
	}

	// We Set even in the case where the source already exists,
	// because we want to update the expiryTime everytime we get the source,
	// as the expiryTime is supposed to reflect the activity (or lack thereof) on that source.
	if err := i.buckets.Set(source, bucket, i.ttl); err != nil {
		return nil, fmt.Errorf("setting buckets: %w", err)
	}

	res := bucket.Reserve()
	if !res.OK() {
		return nil, nil
	}

	delay := res.Delay()
	if delay > i.maxDelay {
		res.Cancel()
	}

	return &delay, nil
}
