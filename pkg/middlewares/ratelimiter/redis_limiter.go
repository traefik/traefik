package ratelimiter

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"golang.org/x/time/rate"
)

type RedisLimiter struct {
	rate    rate.Limit // reqs/s
	burst   int64
	logger  *zerolog.Logger
	limiter *redis_rate.Limiter
	period  ptypes.Duration
}

func NewRedisLimiter(
	rate rate.Limit,
	burst int64,
	config dynamic.RateLimit, logger *zerolog.Logger,
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
	limiter := redis_rate.NewLimiter(rdb)

	return &RedisLimiter{
		rate:    rate,
		burst:   burst,
		period:  config.Period,
		logger:  logger,
		limiter: limiter,
	}, nil
}

func (r *RedisLimiter) Allowed(
	ctx context.Context, source string, _ int64, req *http.Request, rw http.ResponseWriter,
) (bool, error) {
	res, err := r.limiter.Allow(
		ctx,
		source,
		redis_rate.Limit{
			Rate:   int(r.rate),
			Period: time.Duration(r.period),
			Burst:  int(r.burst),
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Could not insert/update bucket")
		http.Error(rw, "could not insert/update bucket", http.StatusInternalServerError)
		return false, err
	}

	if res.Allowed == 0 {
		http.Error(rw, "No bursty traffic allowed", http.StatusTooManyRequests)
		return false, nil
	}

	return true, nil
}
