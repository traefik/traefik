package ratelimiter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/time/rate"
)

func buildRedisLimiter(t *testing.T, redisCfg *dynamic.Redis) *redisLimiter {
	t.Helper()
	logger := zerolog.Nop()
	cfg := dynamic.RateLimit{
		Average: 10,
		Burst:   5,
		Period:  ptypes.Duration(time.Second),
		Redis:   redisCfg,
	}
	l, err := newRedisLimiter(context.Background(), rate.Limit(10), 5, time.Second, 60, cfg, &logger)
	require.NoError(t, err)
	rl, ok := l.(*redisLimiter)
	require.True(t, ok)
	return rl
}

// TestRedisLimiter_SharedClientPerConfig pins the fix for the reload leak:
// every limiter built from the same Redis configuration must share one
// client. Traefik rebuilds middleware instances per router chain on every
// dynamic reload without closing the old clients, so a client per instance
// leaks goroutines and heap for the process lifetime (go-redis >= 9.12
// spawns a maintnotifications cleanup goroutine per client, pinning it
// against GC).
func TestRedisLimiter_SharedClientPerConfig(t *testing.T) {
	cfgA := &dynamic.Redis{Endpoints: []string{"127.0.0.1:6379"}, DB: 0}

	before := len(sharedRedisClients.clients)

	first := buildRedisLimiter(t, cfgA)
	for range 20 {
		l := buildRedisLimiter(t, &dynamic.Redis{Endpoints: []string{"127.0.0.1:6379"}, DB: 0})
		assert.True(t, l.client == first.client, "identical config must reuse the same client instance")
	}
	assert.Equal(t, before+1, len(sharedRedisClients.clients), "21 constructions with one config must add exactly one cached client")
}

// TestRedisLimiter_DistinctClientPerDistinctConfig verifies the cache keys
// on configuration values: a different DB (or any other field) must get its
// own client rather than silently sharing state.
func TestRedisLimiter_DistinctClientPerDistinctConfig(t *testing.T) {
	a := buildRedisLimiter(t, &dynamic.Redis{Endpoints: []string{"127.0.0.1:6379"}, DB: 1})
	b := buildRedisLimiter(t, &dynamic.Redis{Endpoints: []string{"127.0.0.1:6379"}, DB: 2})
	assert.False(t, a.client == b.client, "different config must not share a client")
}

// TestClientTLSFingerprint_TracksFileContent verifies the cache key follows
// the materialized TLS bytes: rotating a certificate at an unchanged file
// path must change the fingerprint, so the cache builds a fresh client
// instead of pinning pre-rotation material for the process lifetime.
func TestClientTLSFingerprint_TracksFileContent(t *testing.T) {
	assert.Empty(t, clientTLSFingerprint(nil))

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(caPath, []byte("cert-generation-1"), 0o600))
	before := clientTLSFingerprint(&types.ClientTLS{CA: caPath})

	require.NoError(t, os.WriteFile(caPath, []byte("cert-generation-2"), 0o600))
	after := clientTLSFingerprint(&types.ClientTLS{CA: caPath})
	assert.NotEqual(t, before, after, "rotated file content at the same path must change the fingerprint")

	// Inline values (non-paths) hash their literal content.
	assert.NotEqual(t,
		clientTLSFingerprint(&types.ClientTLS{CA: "inline-a"}),
		clientTLSFingerprint(&types.ClientTLS{CA: "inline-b"}))
	// InsecureSkipVerify participates in the key.
	assert.NotEqual(t,
		clientTLSFingerprint(&types.ClientTLS{CA: "inline-a"}),
		clientTLSFingerprint(&types.ClientTLS{CA: "inline-a", InsecureSkipVerify: true}))
}
