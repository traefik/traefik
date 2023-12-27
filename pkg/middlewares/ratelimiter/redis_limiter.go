package ratelimiter

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/ratelimiter/redisrate"
	"golang.org/x/time/rate"
)

type RedisLimiter struct {
	rate     rate.Limit // reqs/s
	burst    int64
	maxDelay time.Duration

	period  ptypes.Duration
	logger  *zerolog.Logger
	limiter *redisrate.Limiter
}

func NewRedisLimiter(
	rate rate.Limit,
	burst int64,
	maxDelay time.Duration,
	ttl int,
	config dynamic.RateLimit,
	logger *zerolog.Logger,
) (Limiter, error) {
	options := &redis.UniversalOptions{
		Addrs:        config.Redis.Endpoints,
		Username:     config.Redis.Username,
		Password:     config.Redis.Password,
		DB:           config.Redis.DB,
		PoolSize:     config.Redis.PoolSize,
		MinIdleConns: config.Redis.MinIdleConns,
		ReadTimeout:  config.Redis.ReadTimeout,
		WriteTimeout: config.Redis.WriteTimeout,
		DialTimeout:  config.Redis.DialTimeout,
	}
	if config.Redis.TLS != nil {
		tlsConfig, err := config.Redis.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, err
		}
		options.TLSConfig = tlsConfig
	}
	rdb := redis.NewUniversalClient(options)

	limiter := redisrate.NewLimiter(rdb, ttl, maxDelay)

	return &RedisLimiter{
		rate:     rate,
		burst:    burst,
		period:   config.Period,
		maxDelay: maxDelay,
		logger:   logger,
		limiter:  limiter,
	}, nil
}

func (r *RedisLimiter) Allow(
	ctx context.Context, source string, _ int64, req *http.Request, rw http.ResponseWriter,
) (bool, error) {
	// start := time.Now()
	// // Code to measure
	// defer fmt.Println(measurement(start).Nanoseconds())

	res, err := r.limiter.Allow(
		ctx,
		source,
		redisrate.Limit{
			Rate:   r.rate,
			Period: time.Duration(r.period),
			Burst:  r.burst,
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Could not insert/update bucket")
		http.Error(rw, "could not insert/update bucket", http.StatusInternalServerError)
		return false, err
	}

	if !res.Ok {
		http.Error(rw, "No bursty traffic allowed", http.StatusTooManyRequests)
		return false, nil
	}

	if res.Delay > r.maxDelay {
		r.serveDelayError(ctx, rw, res.Delay)
		return false, nil
	}

	// if res.Delay != 0 {
	// 	fmt.Println("here", res.Delay.Milliseconds(), r.maxDelay.Milliseconds())
	// }
	time.Sleep(res.Delay)

	return true, nil
}

func (r *RedisLimiter) serveDelayError(ctx context.Context, w http.ResponseWriter, delay time.Duration) {
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", math.Ceil(delay.Seconds())))
	w.Header().Set("X-Retry-In", delay.String())
	w.WriteHeader(http.StatusTooManyRequests)

	if _, err := w.Write([]byte(http.StatusText(http.StatusTooManyRequests))); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Could not serve 429")
	}
}
