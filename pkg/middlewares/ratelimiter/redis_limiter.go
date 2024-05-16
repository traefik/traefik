package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
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

	limiter := redisrate.NewLimiter(redis.NewUniversalClient(options), ttl, maxDelay)

	return &RedisLimiter{
		rate:     rate,
		burst:    burst,
		period:   config.Period,
		maxDelay: maxDelay,
		logger:   logger,
		limiter:  limiter,
	}, nil
}

func (r *RedisLimiter) Allow(ctx context.Context, source string) (Result, error) {
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
		return Result{
			Ok: false,
		}, err
	}

	if !res.Ok {
		return Result{
			Ok: false,
		}, nil
	}

	if res.Delay > r.maxDelay {
		return Result{
			Ok:    false,
			Delay: res.Delay,
		}, nil
	}

	time.Sleep(res.Delay)

	return Result{
		Ok:        true,
		Remaining: res.Tokens,
		Delay:     res.Delay,
	}, nil
}
