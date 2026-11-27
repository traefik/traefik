package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/middlewares/ratelimiter/ttlmap"
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
	buckets *ttlmap.Map[*rate.Limiter] // actual buckets, keyed by source.

	logger *zerolog.Logger
}

func newInMemoryRateLimiter(rl rate.Limit, burst int64, maxDelay time.Duration, ttl int, logger *zerolog.Logger) (*inMemoryRateLimiter, error) {
	buckets, err := ttlmap.New[*rate.Limiter](maxSources)
	if err != nil {
		return nil, fmt.Errorf("creating ttlmap: %w", err)
	}

	return &inMemoryRateLimiter{
		rate:     rl,
		burst:    burst,
		maxDelay: maxDelay,
		ttl:      ttl,
		logger:   logger,
		buckets:  buckets,
	}, nil
}

func (i *inMemoryRateLimiter) Allow(_ context.Context, source string) (*time.Duration, error) {
	// The bucket is looked up and, when there is none, created in a single step:
	// otherwise, concurrent first requests from a source would each create their own full bucket.
	// The expiryTime is updated everytime we get the source,
	// as it is supposed to reflect the activity (or lack thereof) on that source.
	bucket, err := i.buckets.GetOrSet(source, func() *rate.Limiter {
		return rate.NewLimiter(i.rate, int(i.burst))
	}, i.ttl)
	if err != nil {
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
