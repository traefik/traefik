package ratelimiter

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/mailgun/ttlmap"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

type BasicLimiter struct {
	rate  rate.Limit // reqs/s
	burst int64
	// maxDelay is the maximum duration we're willing to wait for a bucket reservation to become effective, in nanoseconds.
	// For now it is somewhat arbitrarily set to 1/(2*rate).
	maxDelay time.Duration
	// each rate limiter for a given source is stored in the buckets ttlmap.
	// To keep this ttlmap constrained in size,
	// each ratelimiter is "garbage collected" when it is considered expired.
	// It is considered expired after it hasn't been used for ttl seconds.
	ttl     int
	buckets *ttlmap.TtlMap // actual buckets, keyed by source.

	logger *zerolog.Logger
}

func NewBaseLimiter(
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
	return &BasicLimiter{
		rate:     rate,
		burst:    burst,
		maxDelay: maxDelay,
		ttl:      ttl,
		logger:   logger,

		buckets: buckets,
	}, nil
}

func (b *BasicLimiter) Allow(
	ctx context.Context, source string, amount int64, req *http.Request, rw http.ResponseWriter,
) (bool, error) {
	// Get bucket which contain limiter information.
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
		b.logger.Error().Err(err).Msg("Could not insert/update bucket")
		http.Error(rw, "could not insert/update bucket", http.StatusInternalServerError)
		return false, err
	}

	res := bucket.Reserve()
	if !res.OK() {
		http.Error(rw, "No bursty traffic allowed", http.StatusTooManyRequests)
		return false, nil
	}
	delay := res.Delay()
	if delay > b.maxDelay {
		res.Cancel()
		b.serveDelayError(ctx, rw, delay)
		return false, nil
	}

	time.Sleep(delay)
	return true, nil
}

func (b *BasicLimiter) serveDelayError(ctx context.Context, w http.ResponseWriter, delay time.Duration) {
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", math.Ceil(delay.Seconds())))
	w.Header().Set("X-Retry-In", delay.String())
	w.WriteHeader(http.StatusTooManyRequests)

	if _, err := w.Write([]byte(http.StatusText(http.StatusTooManyRequests))); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Could not serve 429")
	}
}
