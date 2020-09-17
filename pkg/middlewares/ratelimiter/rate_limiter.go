// Package ratelimiter implements a rate limiting and traffic shaping middleware with a set of token buckets.
package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mailgun/ttlmap"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/vulcand/oxy/utils"
	"golang.org/x/time/rate"
)

const (
	typeName   = "RateLimiterType"
	maxSources = 65536
)

// rateLimiter implements rate limiting and traffic shaping with a set of token buckets;
// one for each traffic source. The same parameters are applied to all the buckets.
type rateLimiter struct {
	name  string
	rate  rate.Limit // reqs/s
	burst int64
	// maxDelay is the maximum duration we're willing to wait for a bucket reservation to become effective, in nanoseconds.
	// For now it is somewhat arbitrarily set to 1/(2*rate).
	maxDelay      time.Duration
	sourceMatcher utils.SourceExtractor
	next          http.Handler

	buckets *ttlmap.TtlMap // actual buckets, keyed by source.
}

// New returns a rate limiter middleware.
func New(ctx context.Context, next http.Handler, config dynamic.RateLimit, name string) (http.Handler, error) {
	ctxLog := log.With(ctx, log.Str(log.MiddlewareName, name), log.Str(log.MiddlewareType, typeName))
	log.FromContext(ctxLog).Debug("Creating middleware")

	if config.SourceCriterion == nil ||
		config.SourceCriterion.IPStrategy == nil &&
			config.SourceCriterion.RequestHeaderName == "" && !config.SourceCriterion.RequestHost {
		config.SourceCriterion = &dynamic.SourceCriterion{
			IPStrategy: &dynamic.IPStrategy{},
		}
	}

	sourceMatcher, err := middlewares.GetSourceExtractor(ctxLog, config.SourceCriterion)
	if err != nil {
		return nil, err
	}

	buckets, err := ttlmap.NewConcurrent(maxSources)
	if err != nil {
		return nil, err
	}

	burst := config.Burst
	if burst < 1 {
		burst = 1
	}

	period := time.Duration(config.Period)
	if period == 0 {
		period = time.Second
	}

	// Logically, we should set maxDelay to infinity when config.Average == 0 (because it means no rate limiting),
	// but since the reservation will give us a delay = 0 anyway in this case, we're good even with any maxDelay >= 0.
	var maxDelay time.Duration
	var rtl float64
	if config.Average > 0 {
		rtl = float64(config.Average*int64(time.Second)) / float64(period)
		// maxDelay does not scale well for rates below 1,
		// so we just cap it to the corresponding value, i.e. 0.5s, in order to keep the effective rate predictable.
		// One alternative would be to switch to a no-reservation mode (Allow() method) whenever we are in such a low rate regime.
		if rtl < 1 {
			maxDelay = 500 * time.Millisecond
		} else {
			maxDelay = time.Second / (time.Duration(rtl) * 2)
		}
	}

	return &rateLimiter{
		name:          name,
		rate:          rate.Limit(rtl),
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
	ctx := middlewares.GetLoggerCtx(r.Context(), rl.name, typeName)
	logger := log.FromContext(ctx)

	source, amount, err := rl.sourceMatcher.Extract(r)
	if err != nil {
		logger.Errorf("could not extract source of request: %v", err)
		http.Error(w, "could not extract source of request", http.StatusInternalServerError)
		return
	}

	if amount != 1 {
		logger.Infof("ignoring token bucket amount > 1: %d", amount)
	}

	var bucket *rate.Limiter
	if rlSource, exists := rl.buckets.Get(source); exists {
		bucket = rlSource.(*rate.Limiter)
	} else {
		bucket = rate.NewLimiter(rl.rate, int(rl.burst))
		if err := rl.buckets.Set(source, bucket, int(rl.maxDelay)*10+1); err != nil {
			logger.Errorf("could not insert bucket: %v", err)
			http.Error(w, "could not insert bucket", http.StatusInternalServerError)
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
		rl.serveDelayError(ctx, w, r, delay)
		return
	}

	time.Sleep(delay)
	rl.next.ServeHTTP(w, r)
}

func (rl *rateLimiter) serveDelayError(ctx context.Context, w http.ResponseWriter, r *http.Request, delay time.Duration) {
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", delay.Seconds()))
	w.Header().Set("X-Retry-In", delay.String())
	w.WriteHeader(http.StatusTooManyRequests)

	if _, err := w.Write([]byte(http.StatusText(http.StatusTooManyRequests))); err != nil {
		log.FromContext(ctx).Errorf("could not serve 429: %v", err)
	}
}
