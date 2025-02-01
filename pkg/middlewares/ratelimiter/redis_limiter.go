package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"golang.org/x/time/rate"
)

const redisPrefix = "rate:"

type RedisLimiter struct {
	rate     rate.Limit // reqs/s
	burst    int64
	maxDelay time.Duration

	period  ptypes.Duration
	logger  *zerolog.Logger
	ttl     int
	rClient Rediser
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
		Addrs:          config.Redis.Endpoints,
		Username:       config.Redis.Username,
		Password:       config.Redis.Password,
		DB:             config.Redis.DB,
		PoolSize:       config.Redis.PoolSize,
		MinIdleConns:   config.Redis.MinIdleConns,
		ReadTimeout:    config.Redis.ReadTimeout,
		WriteTimeout:   config.Redis.WriteTimeout,
		DialTimeout:    config.Redis.DialTimeout,
		MaxActiveConns: config.Redis.MaxActiveConns,
	}
	if config.Redis.TLS != nil {
		tlsConfig, err := config.Redis.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, err
		}
		options.TLSConfig = tlsConfig
	}
	rClient := redis.NewUniversalClient(options)
	limiter := &RedisLimiter{
		rate:     rate,
		burst:    burst,
		period:   config.Period,
		maxDelay: maxDelay,
		logger:   logger,
		ttl:      ttl,
		rClient:  rClient,
	}

	return limiter, nil
}

func injectClient(r *RedisLimiter, client Rediser) *RedisLimiter {
	r.rClient = client
	return r
}

func (r *RedisLimiter) Allow(ctx context.Context, source string) (Result, error) {
	res, err := r.evaluateScript(ctx, source)
	if err != nil {
		return Result{}, err
	}

	if !res.Ok {
		return Result{Ok: false}, nil
	}

	if res.Delay > r.maxDelay {
		return Result{
			Ok:    false,
			Delay: res.Delay,
		}, nil
	}

	select {
	case <-ctx.Done():
		return Result{Ok: false}, nil
	case <-time.After(res.Delay):
	}

	return Result{
		Ok:        true,
		Remaining: res.Remaining,
		Delay:     res.Delay,
	}, nil
}

func (r *RedisLimiter) evaluateScript(ctx context.Context, key string) (*Result, error) {
	if r.rate == rate.Inf {
		return &Result{
			Ok:        true,
			Remaining: 1.0,
		}, nil
	}

	rate := r.rate / 1000000
	t := time.Now().UnixMicro()
	params := []interface{}{float64(rate), r.burst, r.ttl, t, r.maxDelay.Microseconds()}
	v, err := AllowTokenBucketScript.Run(ctx, r.rClient, []string{redisPrefix + key}, params...).Result()
	if err != nil {
		return nil, err
	}
	values := v.([]interface{})

	ok, err := strconv.ParseBool(values[0].(string))
	if err != nil {
		return nil, fmt.Errorf("redisrate: fail to parse ok value from redis rate lua script: %w", err)
	}
	delay, err := strconv.ParseFloat(values[1].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("redisrate: fail to parse delay value from redis rate lua script: %w", err)
	}
	tokens, err := strconv.ParseFloat(values[2].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("redisrate: fail to parse tokens value from redis rate lua script: %w", err)
	}

	res := &Result{
		Ok:        ok,
		Remaining: tokens,
		Delay:     time.Duration(delay * float64(time.Microsecond)),
	}
	return res, nil
}
