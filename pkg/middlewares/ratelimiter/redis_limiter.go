package ratelimiter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktypes "github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/time/rate"
)

const redisPrefix = "rate:"

// sharedRedisClients caches one client per distinct Redis configuration.
//
// Traefik rebuilds every middleware instance per router chain on every
// dynamic configuration reload, and the dropped instances are never told to
// close their clients. A fresh client per instance therefore leaks: since
// go-redis v9.12 each NewUniversalClient (RESP3 + default "auto"
// maintnotifications) spawns a CircuitBreakerManager.cleanupLoop goroutine
// at construction, and a live goroutine is a GC root — orphaned clients are
// pinned forever. The leak scales with routers-referencing-the-middleware
// times reload frequency, each orphan holding a goroutine and its managers
// until the process hits its memory ceiling.
//
// Sharing one client per unique configuration keeps the client count
// bounded regardless of router fan-out or reload frequency. Clients live
// for the process lifetime, which matches how the previous per-instance
// clients behaved in aggregate (none were ever closed).
var sharedRedisClients = struct {
	mu      sync.Mutex
	clients map[string]redis.UniversalClient
}{clients: make(map[string]redis.UniversalClient)}

// sharedRedisClient returns the cached client for the given Redis
// configuration, creating it on first use. The cache key is the JSON
// serialization of the dynamic config (pointers dereferenced, values
// compared) plus a fingerprint of the materialized TLS inputs — never the
// options struct, whose function and TLS pointers would differ on every
// reload and defeat the cache.
func sharedRedisClient(config *dynamic.Redis, options *redis.UniversalOptions) (redis.UniversalClient, error) {
	rawKey, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshaling redis config for client cache key: %w", err)
	}
	key := string(rawKey) + "|" + clientTLSFingerprint(config.TLS)

	sharedRedisClients.mu.Lock()
	defer sharedRedisClients.mu.Unlock()

	if client, ok := sharedRedisClients.clients[key]; ok {
		return client, nil
	}
	client := redis.NewUniversalClient(options)
	sharedRedisClients.clients[key] = client
	return client, nil
}

// clientTLSFingerprint hashes the materialized TLS inputs: when CA/Cert/Key
// are file paths the file contents are hashed, otherwise the literal
// (inline PEM) values are. Certificates rotated at an unchanged path
// therefore produce a new cache key and a fresh client, instead of the
// cache pinning the pre-rotation material for the process lifetime.
func clientTLSFingerprint(clientTLS *traefiktypes.ClientTLS) string {
	if clientTLS == nil {
		return ""
	}
	h := sha256.New()
	for _, v := range []string{clientTLS.CA, clientTLS.Cert, clientTLS.Key} {
		if content, err := os.ReadFile(v); err == nil {
			h.Write(content)
		} else {
			h.Write([]byte(v))
		}
		h.Write([]byte{0})
	}
	if clientTLS.InsecureSkipVerify {
		h.Write([]byte{1})
	}
	return hex.EncodeToString(h.Sum(nil))
}

type redisLimiter struct {
	rate     rate.Limit // reqs/s
	burst    int64
	maxDelay time.Duration
	period   ptypes.Duration
	logger   *zerolog.Logger
	ttl      int
	client   Rediser
	script   *redis.Script
}

func newRedisLimiter(ctx context.Context, rate rate.Limit, burst int64, maxDelay time.Duration, ttl int, config dynamic.RateLimit, logger *zerolog.Logger) (limiter, error) {
	options := &redis.UniversalOptions{
		Addrs:          config.Redis.Endpoints,
		Username:       config.Redis.Username,
		Password:       config.Redis.Password,
		DB:             config.Redis.DB,
		PoolSize:       config.Redis.PoolSize,
		MinIdleConns:   config.Redis.MinIdleConns,
		MaxActiveConns: config.Redis.MaxActiveConns,
	}

	if config.Redis.DialTimeout != nil && *config.Redis.DialTimeout > 0 {
		options.DialTimeout = time.Duration(*config.Redis.DialTimeout)
	}

	if config.Redis.ReadTimeout != nil {
		if *config.Redis.ReadTimeout > 0 {
			options.ReadTimeout = time.Duration(*config.Redis.ReadTimeout)
		} else {
			options.ReadTimeout = -1
		}
	}

	if config.Redis.WriteTimeout != nil {
		if *config.Redis.WriteTimeout > 0 {
			options.WriteTimeout = time.Duration(*config.Redis.WriteTimeout)
		} else {
			options.WriteTimeout = -1
		}
	}

	if config.Redis.TLS != nil {
		var err error
		options.TLSConfig, err = config.Redis.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating TLS config: %w", err)
		}
	}

	script, err := LoadTokenBucketScript()
	if err != nil {
		return nil, fmt.Errorf("loading bucket script: %w", err)
	}

	client, err := sharedRedisClient(config.Redis, options)
	if err != nil {
		return nil, fmt.Errorf("obtaining redis client: %w", err)
	}

	return &redisLimiter{
		rate:     rate,
		burst:    burst,
		period:   config.Period,
		maxDelay: maxDelay,
		logger:   logger,
		ttl:      ttl,
		client:   client,
		script:   script,
	}, nil
}

func (r *redisLimiter) Allow(ctx context.Context, source string) (*time.Duration, error) {
	ok, delay, err := r.evaluateScript(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("evaluating script: %w", err)
	}
	if !ok {
		return nil, nil
	}
	return delay, nil
}

func (r *redisLimiter) evaluateScript(ctx context.Context, key string) (bool, *time.Duration, error) {
	if r.rate == rate.Inf {
		return true, nil, nil
	}

	params := []any{
		float64(r.rate / 1000000),
		r.burst,
		r.ttl,
		time.Now().UnixMicro(),
		r.maxDelay.Microseconds(),
	}
	v, err := r.script.Run(ctx, r.client, []string{redisPrefix + key}, params...).Result()
	if err != nil {
		return false, nil, fmt.Errorf("running script: %w", err)
	}

	values := v.([]any)
	ok, err := strconv.ParseBool(values[0].(string))
	if err != nil {
		return false, nil, fmt.Errorf("parsing ok value from redis rate lua script: %w", err)
	}
	delay, err := strconv.ParseFloat(values[1].(string), 64)
	if err != nil {
		return false, nil, fmt.Errorf("parsing delay value from redis rate lua script: %w", err)
	}

	microDelay := time.Duration(delay * float64(time.Microsecond))
	return ok, &microDelay, nil
}
