// Package ratelimiter implements a rate limiting and traffic shaping middleware with a set of token buckets.
package ratelimiter

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/vulcand/oxy/v2/utils"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
)

const (
	typeName   = "RateLimiter"
	maxSources = 65536
)

// rateLimiter implements rate limiting and traffic shaping with a set of token buckets;
// one for each traffic source. The same parameters are applied to all the buckets.
type rateLimiter struct {
	name string
	rate rate.Limit // reqs/s
	// maxDelay is the maximum duration we're willing to wait for a bucket reservation to become effective, in nanoseconds.
	// For now it is somewhat arbitrarily set to 1/(2*rate).
	maxDelay      time.Duration
	sourceMatcher utils.SourceExtractor
	next          http.Handler
	logger        *zerolog.Logger

	limiter Limiter
}

type Result struct {
	Ok        bool
	Remaining float64
	Delay     time.Duration
}

type Limiter interface {
	Allow(ctx context.Context, token string) (Result, error)
}

// New returns a rate limiter middleware.
func New(ctx context.Context, next http.Handler, config dynamic.RateLimit, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	ctxLog := logger.WithContext(ctx)

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

	burst := config.Burst
	if burst < 1 {
		burst = 1
	}

	period := time.Duration(config.Period)
	if period < 0 {
		return nil, fmt.Errorf("negative value not valid for period: %v", period)
	}
	if period == 0 {
		period = time.Second
	}

	// Initialized at rate.Inf to enforce no rate limiting when config.Average == 0
	rtl := float64(rate.Inf)
	// No need to set any particular value for maxDelay as the reservation's delay
	// will be <= 0 in the Inf case (i.e. the average == 0 case).
	var maxDelay time.Duration

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

	// Make the ttl inversely proportional to how often a rate limiter is supposed to see any activity (when maxed out),
	// for low rate limiters.
	// Otherwise just make it a second for all the high rate limiters.
	// Add an extra second in both cases for continuity between the two cases.
	ttl := 1
	if rtl >= 1 {
		ttl++
	} else if rtl > 0 {
		ttl += int(1 / rtl)
	}
	var limiter Limiter
	if config.Redis != nil {
		limiter, err = NewRedisLimiter(rate.Limit(rtl), burst, maxDelay, ttl, config, logger)
		if err != nil {
			return nil, err
		}
	} else {
		limiter, err = NewInMemoryRateLimiter(rate.Limit(rtl), burst, maxDelay, ttl, logger)
		if err != nil {
			return nil, err
		}
	}

	rateLimmiter := &rateLimiter{
		logger:        logger,
		name:          name,
		rate:          rate.Limit(rtl),
		maxDelay:      maxDelay,
		next:          next,
		sourceMatcher: sourceMatcher,
		limiter:       limiter,
	}

	return rateLimmiter, nil
}

func (rl *rateLimiter) GetTracingInformation() (string, string, trace.SpanKind) {
	return rl.name, typeName, trace.SpanKindInternal
}

func (rl *rateLimiter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), rl.name, typeName)
	ctx := logger.WithContext(req.Context())

	source, amount, err := rl.sourceMatcher.Extract(req)
	if err != nil {
		logger.Error().Err(err).Msg("Could not extract source of request")
		http.Error(rw, "could not extract source of request", http.StatusInternalServerError)
		return
	}

	// always equal to 1 if exist.
	if amount != 1 {
		logger.Info().Msgf("ignoring token bucket amount > 1: %d", amount)
	}

	result, err := rl.limiter.Allow(ctx, source)
	if err != nil {
		rl.logger.Error().Err(err).Msg("Could not insert/update bucket")
		observability.SetStatusErrorf(ctx, "Could not insert/update bucket")
		http.Error(rw, "could not insert/update bucket", http.StatusInternalServerError)
		return
	}

	if !result.Ok {
		if result.Delay > rl.maxDelay {
			rl.serveDelayError(ctx, rw, result.Delay)
		} else {
			observability.SetStatusErrorf(ctx, "No bursty traffic allowed")
			http.Error(rw, "No bursty traffic allowed", http.StatusTooManyRequests)
		}
	}

	if result.Ok {
		rl.next.ServeHTTP(rw, req)
	}
}

func (rl *rateLimiter) serveDelayError(ctx context.Context, w http.ResponseWriter, delay time.Duration) {
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", math.Ceil(delay.Seconds())))
	w.Header().Set("X-Retry-In", delay.String())
	w.WriteHeader(http.StatusTooManyRequests)

	if _, err := w.Write([]byte(http.StatusText(http.StatusTooManyRequests))); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Could not serve 429")
	}
}
