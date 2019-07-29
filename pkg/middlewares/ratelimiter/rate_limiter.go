// Package ratelimiter implements a rate limiting and traffic shaping middleware
// with a set of token buckets.
package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/mailgun/ttlmap"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/vulcand/oxy/utils"
	"golang.org/x/time/rate"
)

const (
	typeName   = "RateLimiterType"
	maxSources = 65536
)

// rateLimiter implements rate limiting and traffic shaping with a set of token
// buckets; one for each traffic source. The same parameters are applied to all
// the buckets.
type rateLimiter struct {
	name  string
	rate  rate.Limit // reqs/s
	burst int64
	// maxDelay is the maximum duration we're willing to wait for a bucket reservation
	// to become effective, in nanoseconds.
	// For now it is somewhat arbitrarily set to 1/rate.
	maxDelay      time.Duration
	sourceMatcher utils.SourceExtractor
	next          http.Handler

	bucketsMu sync.Mutex
	buckets   *ttlmap.TtlMap // actual buckets, keyed by source.
}

// New returns a rate limiter middleware.
func New(ctx context.Context, next http.Handler, config dynamic.RateLimit, name string) (http.Handler, error) {
	logctx := log.With(ctx, log.Str(log.MiddlewareName, name), log.Str(log.MiddlewareType, typeName))
	log.FromContext(logctx).Debug("Creating middleware")

	if config.SourceCriterion == nil ||
		config.SourceCriterion.IPStrategy == nil &&
			config.SourceCriterion.RequestHeaderName == "" && !config.SourceCriterion.RequestHost {
		config.SourceCriterion = &dynamic.SourceCriterion{
			IPStrategy: &dynamic.IPStrategy{},
		}
	}

	sourceMatcher, err := middlewares.GetSourceExtractor(logctx, config.SourceCriterion)
	if err != nil {
		return nil, err
	}

	buckets, err := ttlmap.NewMap(maxSources)
	if err != nil {
		return nil, err
	}

	burst := config.Burst
	if burst <= 0 {
		burst = 1
	}

	// Logically, we should set maxDelay to ~infinity when config.Average == 0
	// (because it means to rate limiting), but since the reservation will give us a
	// delay = 0 anyway in this case, we're good even with any maxDelay >= 0.
	var maxDelay time.Duration
	if config.Average != 0 {
		maxDelay = time.Second / time.Duration(config.Average*2)
	}

	return &rateLimiter{
		name:          name,
		rate:          rate.Limit(config.Average),
		burst:         burst,
		maxDelay:      maxDelay,
		next:          next,
		sourceMatcher: sourceMatcher,
		buckets:       buckets,
	}, nil
}

func (rl *rateLimiter) GetTracingInformation() (string, ext.SpanKindEnum) {
	return rl.name, tracing.SpanKindNoneEnum
}

func (rl *rateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	source, amount, err := rl.sourceMatcher.Extract(r)
	if err != nil {
		middlewares.GetLogger(r.Context(), rl.name, typeName).Errorf("could not extract source of request: %v", err)
		http.Error(w, "could not extract source of request", 500)
		return
	}
	if amount != 1 {
		middlewares.GetLogger(r.Context(), rl.name, typeName).Infof("ignoring token bucket amount > 1: %d", amount)
	}

	rl.bucketsMu.Lock()
	defer rl.bucketsMu.Unlock()

	rateLimiterI, exists := rl.buckets.Get(source)
	var bucket *rate.Limiter
	if exists {
		bucket = rateLimiterI.(*rate.Limiter)
	} else {
		bucket = rate.NewLimiter(rl.rate, int(rl.burst))
		if err := rl.buckets.Set(source, bucket, int(rl.maxDelay)*10+1); err != nil {
			middlewares.GetLogger(r.Context(), rl.name, typeName).Errorf("could not insert bucket: %v", err)
			http.Error(w, "could not insert bucket", 500)
			return
		}
	}

	res := bucket.Reserve()
	if !res.OK() {
		http.Error(w, "No bursty traffic allowed", http.StatusTooManyRequests)
		return
	}
	delay := res.Delay()
	if delay > rl.maxDelay {
		res.Cancel()
		rl.serveDelayError(w, r, delay)
		return
	}
	time.Sleep(delay)
	rl.next.ServeHTTP(w, r)
}

func (rl *rateLimiter) serveDelayError(w http.ResponseWriter, r *http.Request, delay time.Duration) {
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", delay.Seconds()))
	w.Header().Set("X-Retry-In", delay.String())
	w.WriteHeader(http.StatusTooManyRequests)
	if _, err := w.Write([]byte(http.StatusText(http.StatusTooManyRequests))); err != nil {
		middlewares.GetLogger(r.Context(), rl.name, typeName).Errorf("could not serve 429: %v", err)
	}
}
