package confighash

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tls"
)

// minimalDynamicConfig returns a hand-built *dynamic.Configuration that
// exercises HTTP, TCP, UDP and TLS sub-trees with a few representative
// pointer / map / slice fields. Used as the seed for sensitivity and
// order-independence tests.
func minimalDynamicConfig() *dynamic.Configuration {
	passHostHeader := true

	return &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
				"r1": {
					EntryPoints: []string{"web"},
					Middlewares: []string{"m1"},
					Service:     "s1",
					Rule:        "Host(`example.com`)",
					Priority:    10,
				},
			},
			Services: map[string]*dynamic.Service{
				"s1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers:        []dynamic.Server{{URL: "http://10.0.0.1:80"}},
						PassHostHeader: &passHostHeader,
					},
				},
			},
			Middlewares: map[string]*dynamic.Middleware{
				"m1": {
					AddPrefix: &dynamic.AddPrefix{Prefix: "/api"},
				},
			},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers: map[string]*dynamic.TCPRouter{
				"t1": {
					EntryPoints: []string{"tcp-ep"},
					Service:     "ts1",
					Rule:        "HostSNI(`*`)",
				},
			},
			Services: map[string]*dynamic.TCPService{
				"ts1": {
					LoadBalancer: &dynamic.TCPServersLoadBalancer{
						Servers: []dynamic.TCPServer{{Address: "10.0.0.1:443"}},
					},
				},
			},
		},
		UDP: &dynamic.UDPConfiguration{
			Routers: map[string]*dynamic.UDPRouter{
				"u1": {
					EntryPoints: []string{"udp-ep"},
					Service:     "us1",
				},
			},
			Services: map[string]*dynamic.UDPService{
				"us1": {
					LoadBalancer: &dynamic.UDPServersLoadBalancer{
						Servers: []dynamic.UDPServer{{Address: "10.0.0.1:53"}},
					},
				},
			},
		},
		TLS: &dynamic.TLSConfiguration{
			Options: map[string]tls.Options{
				"default": {MinVersion: "VersionTLS12"},
			},
		},
	}
}

// Contract 1 — Determinism.

func TestHash_DeterministicAcrossCalls(t *testing.T) {
	t.Parallel()

	cfg := minimalDynamicConfig()
	want := Hash(cfg)
	for i := 0; i < 100; i++ {
		assert.Equal(t, want, Hash(cfg), "iteration %d", i)
	}
}

func TestHash_DeterministicAcrossCopies(t *testing.T) {
	t.Parallel()

	cfg := &dynamic.Configuration{}
	_, err := toml.DecodeFile("../fixtures/sample.toml", cfg)
	require.NoError(t, err)

	cpy := cfg.DeepCopy()
	assert.Equal(t, Hash(cfg), Hash(cpy))
}

// Contract 2 — Sensitivity. Any structural change flips the hash.

func TestHash_DetectsMutation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*dynamic.Configuration)
	}{
		{
			name: "router rule changes",
			mutate: func(c *dynamic.Configuration) {
				c.HTTP.Routers["r1"].Rule = "Host(`other.com`)"
			},
		},
		{
			name: "router priority changes",
			mutate: func(c *dynamic.Configuration) {
				c.HTTP.Routers["r1"].Priority = 20
			},
		},
		{
			name: "router middlewares slice grows",
			mutate: func(c *dynamic.Configuration) {
				c.HTTP.Routers["r1"].Middlewares = append(c.HTTP.Routers["r1"].Middlewares, "m2")
			},
		},
		{
			name: "service PassHostHeader pointer flips",
			mutate: func(c *dynamic.Configuration) {
				b := false
				c.HTTP.Services["s1"].LoadBalancer.PassHostHeader = &b
			},
		},
		{
			name: "service PassHostHeader becomes nil",
			mutate: func(c *dynamic.Configuration) {
				c.HTTP.Services["s1"].LoadBalancer.PassHostHeader = nil
			},
		},
		{
			name: "middleware swapped to a different middleware",
			mutate: func(c *dynamic.Configuration) {
				c.HTTP.Middlewares["m1"] = &dynamic.Middleware{
					StripPrefix: &dynamic.StripPrefix{Prefixes: []string{"/api"}},
				}
			},
		},
		{
			name: "tcp router rule changes",
			mutate: func(c *dynamic.Configuration) {
				c.TCP.Routers["t1"].Rule = "HostSNI(`example.com`)"
			},
		},
		{
			name: "udp router entry points change",
			mutate: func(c *dynamic.Configuration) {
				c.UDP.Routers["u1"].EntryPoints = []string{"udp-ep", "udp-ep-2"}
			},
		},
		{
			name: "tls option field changes",
			mutate: func(c *dynamic.Configuration) {
				opt := c.TLS.Options["default"]
				opt.MinVersion = "VersionTLS13"
				c.TLS.Options["default"] = opt
			},
		},
		{
			name: "map gains an entry",
			mutate: func(c *dynamic.Configuration) {
				c.HTTP.Routers["r2"] = &dynamic.Router{Rule: "Host(`b.com`)"}
			},
		},
		{
			name: "map loses an entry",
			mutate: func(c *dynamic.Configuration) {
				delete(c.HTTP.Routers, "r1")
			},
		},
		{
			name: "server slice element changes",
			mutate: func(c *dynamic.Configuration) {
				c.HTTP.Services["s1"].LoadBalancer.Servers[0].URL = "http://10.0.0.2:80"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			base := minimalDynamicConfig()
			want := Hash(base)

			mutated := minimalDynamicConfig()
			test.mutate(mutated)

			assert.NotEqual(t, want, Hash(mutated),
				"mutation %q produced an identical hash", test.name)
		})
	}
}

func TestHash_NilEmptyZeroDistinct(t *testing.T) {
	t.Parallel()

	// Each pair must hash to a different value. Picking pairs that share
	// either the bit-pattern or the underlying zero value but differ in
	// kind, nil-ness or container shape.
	type pair struct {
		name string
		a, b any
	}

	zeroInt := 0
	zeroPtrInt := &zeroInt

	pairs := []pair{
		{"untyped nil vs empty struct ptr", nil, &dynamic.Configuration{}},
		{"nil slice vs empty slice", []string(nil), []string{}},
		{"nil map vs empty map", map[string]int(nil), map[string]int{}},
		{"empty map vs single zero-key entry", map[string]int{}, map[string]int{"": 0}},
		{"empty string vs empty bytes", "", []byte{}},
		{"int64(0) vs uint64(0)", int64(0), uint64(0)},
		{"int64(0) vs bool false", int64(0), false},
		{"uint64(0) vs bool false", uint64(0), false},
		{"int(0) vs *int(0)", 0, zeroPtrInt},
		{"nil ptr vs &T{}", (*dynamic.Configuration)(nil), &dynamic.Configuration{}},
	}

	for _, p := range pairs {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			assert.NotEqual(t, Hash(p.a), Hash(p.b),
				"hash collision between %#v and %#v", p.a, p.b)
		})
	}
}

// Contract 3 — Map-order independence (the XOR design's whole point).

func TestHash_MapOrderIndependent(t *testing.T) {
	t.Parallel()

	// Primitive-value map.
	keys := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"}
	want := Hash(buildIntMap(keys))
	// Reverse, rotate, and shuffle insertion order; the hash must not move.
	orders := [][]string{
		reversed(keys),
		rotated(keys, 3),
		{"theta", "alpha", "eta", "beta", "zeta", "gamma", "epsilon", "delta"},
		{"epsilon", "delta", "alpha", "gamma", "zeta", "beta", "eta", "theta"},
	}
	for i, ord := range orders {
		assert.Equal(t, want, Hash(buildIntMap(ord)), "primitive map order %d", i)
	}

	// Pointer-value map (the realistic case — routers/services/middlewares
	// are all map[string]*T).
	routerKeys := []string{"r1", "r2", "r3", "r4", "r5", "r6"}
	wantR := Hash(buildRouterMap(routerKeys))
	for i, ord := range [][]string{
		reversed(routerKeys),
		rotated(routerKeys, 2),
		{"r4", "r1", "r6", "r2", "r5", "r3"},
	} {
		assert.Equal(t, wantR, Hash(buildRouterMap(ord)), "router map order %d", i)
	}
}

func TestHash_NestedMapOrderIndependent(t *testing.T) {
	t.Parallel()

	keys := []string{"a", "b", "c", "d", "e", "f", "g"}

	cfg1 := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     buildRouterMap(keys),
			Middlewares: buildMiddlewareMap(keys),
		},
	}
	cfg2 := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     buildRouterMap(reversed(keys)),
			Middlewares: buildMiddlewareMap(reversed(keys)),
		},
	}

	assert.Equal(t, Hash(cfg1), Hash(cfg2))
}

// Contract 4 — Realistic configs hash and stay stable across DeepCopy,
// then flip on any mutation deep inside the tree.

func TestHash_SampleFixture(t *testing.T) {
	t.Parallel()

	cfg := &dynamic.Configuration{}
	_, err := toml.DecodeFile("../fixtures/sample.toml", cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTP)
	require.NotEmpty(t, cfg.HTTP.Routers, "fixture must populate at least one router")

	before := Hash(cfg)

	// Deep-copy must hash identically — this is the closest in-process
	// analogue to the provider rebuilding *dynamic.Configuration on every
	// event.
	copied := cfg.DeepCopy()
	assert.Equal(t, before, Hash(copied),
		"deep copy should hash identically to the original")

	// Mutating a leaf string deep inside HTTP.Routers must flip the hash.
	for k := range cfg.HTTP.Routers {
		cfg.HTTP.Routers[k].Rule = "mutated"
		break
	}
	assert.NotEqual(t, before, Hash(cfg),
		"mutating a nested router rule should change the hash")
}

// (A dedicated TestHash_StaticConfig was considered but dropped because
// pkg/config/static transitively imports pkg/config/dynamic/confighash via
// the CRD provider, which would create a test-time import cycle. The
// dynamic-configuration cases above already exercise every kind the
// reflection walker needs to handle.)

// Contract 5 — Resilience: never panics, race-safe.

func TestHash_DoesNotPanicOnUnsupportedKinds(t *testing.T) {
	t.Parallel()

	type weird struct {
		Ch      chan int
		Fn      func()
		Unsafe  unsafe.Pointer
		Payload string
	}

	v := weird{
		Ch:      make(chan int),
		Fn:      func() {},
		Unsafe:  unsafe.Pointer(t),
		Payload: "ok",
	}

	assert.NotPanics(t, func() { _ = Hash(v) })
	// Two values identical-except-for-the-Payload must still hash differently,
	// proving the unsupported branches don't swallow the payload field.
	v2 := v
	v2.Payload = "different"
	assert.NotEqual(t, Hash(v), Hash(v2))
}

func TestHash_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	cfg := minimalDynamicConfig()
	want := Hash(cfg)

	const goroutines = 200
	const calls = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]uint64, goroutines*calls)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for c := 0; c < calls; c++ {
				results[g*calls+c] = Hash(cfg)
			}
		}(g)
	}
	wg.Wait()

	for i, got := range results {
		require.Equalf(t, want, got, "goroutine result %d diverged", i)
	}
}

// --- helpers ---

// buildIntMap derives the value from the key so the result depends only on
// the set of keys, not on insertion order.
func buildIntMap(keys []string) map[string]int {
	m := make(map[string]int, len(keys))
	for _, k := range keys {
		m[k] = len(k)
	}
	return m
}

// buildRouterMap is order-stable for the same reason as buildIntMap.
func buildRouterMap(keys []string) map[string]*dynamic.Router {
	m := make(map[string]*dynamic.Router, len(keys))
	for _, k := range keys {
		m[k] = &dynamic.Router{
			Service:  "svc-" + k,
			Rule:     "Host(`" + k + ".example.com`)",
			Priority: len(k),
		}
	}
	return m
}

func buildMiddlewareMap(keys []string) map[string]*dynamic.Middleware {
	m := make(map[string]*dynamic.Middleware, len(keys))
	for _, k := range keys {
		m[k] = &dynamic.Middleware{
			AddPrefix: &dynamic.AddPrefix{Prefix: "/" + k},
		}
	}
	return m
}

func reversed(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[len(in)-1-i] = v
	}
	return out
}

func rotated(in []string, n int) []string {
	if len(in) == 0 {
		return in
	}
	n %= len(in)
	out := make([]string, 0, len(in))
	out = append(out, in[n:]...)
	out = append(out, in[:n]...)
	return out
}
