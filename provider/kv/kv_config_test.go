package kv

import (
	"sort"
	"strconv"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
)

func aKVPair(key string, value string) *store.KVPair {
	return &store.KVPair{Key: key, Value: []byte(value)}
}

func TestProviderBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		expected *types.Configuration
	}{
		{
			desc: "name with dot",
			kvPairs: filler("traefik",
				frontend("frontend.with.dot",
					withPair("backend", "backend.with.dot.too"),
					withPair("routes/route.with.dot/rule", "Host:test.localhost")),
				backend("backend.with.dot.too",
					withPair("servers/server.with.dot/url", "http://172.17.0.2:80"),
					withPair("servers/server.with.dot/weight", "0"),
					withPair("servers/server.with.dot.without.url/weight", "0")),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend.with.dot.too": {
						Servers: map[string]types.Server{
							"server.with.dot": {
								URL:    "http://172.17.0.2:80",
								Weight: 0,
							},
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend.with.dot": {
						Backend:        "backend.with.dot.too",
						PassHostHeader: true,
						EntryPoints:    []string{},
						Routes: map[string]types.Route{
							"route.with.dot": {
								Rule: "Host:test.localhost",
							},
						},
					},
				},
			},
		},
		{
			desc: "all parameters",
			kvPairs: filler("traefik",
				backend("backend1",
					withPair(pathBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
					withPair(pathBackendLoadBalancerMethod, "drr"),
					withPair(pathBackendLoadBalancerSticky, "true"),
					withPair(pathBackendLoadBalancerStickiness, "true"),
					withPair(pathBackendLoadBalancerStickinessCookieName, "tomate"),
					withPair(pathBackendHealthCheckPath, "/health"),
					withPair(pathBackendHealthCheckPort, "80"),
					withPair(pathBackendHealthCheckInterval, "30s"),
					withPair(pathBackendMaxConnAmount, "5"),
					withPair(pathBackendMaxConnExtractorFunc, "client.ip"),
					withPair("servers/server1/url", "http://172.17.0.2:80"),
					withPair("servers/server1/weight", "0"),
					withPair("servers/server2/weight", "0")),
				frontend("frontend1",
					withPair(pathFrontendBackend, "backend1"),
					withPair(pathFrontendPriority, "6"),
					withPair(pathFrontendPassHostHeader, "false"),
					withPair(pathFrontendEntryPoints, "http,https"),
					withPair("routes/route1/rule", "Host:test.localhost"),
					withPair("routes/route2/rule", "Path:/foo")),
				entry("tlsconfiguration/foo",
					withPair("entrypoints", "http,https"),
					withPair("certificate/certfile", "certfile1"),
					withPair("certificate/keyfile", "keyfile1")),
				entry("tlsconfiguration/bar",
					withPair("entrypoints", "http,https"),
					withPair("certificate/certfile", "certfile2"),
					withPair("certificate/keyfile", "keyfile2")),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend1": {
						Servers: map[string]types.Server{
							"server1": {
								URL:    "http://172.17.0.2:80",
								Weight: 0,
							},
						},
						CircuitBreaker: &types.CircuitBreaker{
							Expression: "NetworkErrorRatio() > 1",
						},
						LoadBalancer: &types.LoadBalancer{
							Method: "drr",
							Sticky: true,
							Stickiness: &types.Stickiness{
								CookieName: "tomate",
							},
						},
						MaxConn: &types.MaxConn{
							Amount:        5,
							ExtractorFunc: "client.ip",
						},
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Port:     0,
							Interval: "30s",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend1": {
						Priority:    6,
						EntryPoints: []string{"http", "https"},
						Backend:     "backend1",
						Routes: map[string]types.Route{
							"route1": {
								Rule: "Host:test.localhost",
							},
							"route2": {
								Rule: "Path:/foo",
							},
						},
					},
				},
				TLSConfiguration: []*tls.Configuration{
					{
						EntryPoints: []string{"http", "https"},
						Certificate: &tls.Certificate{
							CertFile: "certfile2",
							KeyFile:  "keyfile2",
						},
					},
					{
						EntryPoints: []string{"http", "https"},
						Certificate: &tls.Certificate{
							CertFile: "certfile1",
							KeyFile:  "keyfile1",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				Prefix: "traefik",
				kvClient: &Mock{
					KVPairs: test.kvPairs,
				},
			}

			actual := p.buildConfiguration()
			assert.NotNil(t, actual)

			assert.EqualValues(t, test.expected.Backends, actual.Backends)
			assert.EqualValues(t, test.expected.Frontends, actual.Frontends)
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestProviderList(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected []string
	}{
		{
			desc:     "empty key parts and empty store",
			keyParts: []string{},
			expected: []string{},
		},
		{
			desc:     "when non existing key and empty store",
			keyParts: []string{"traefik"},
			expected: []string{},
		},
		{
			desc: "when non existing key",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"bar"},
			expected: []string{},
		},
		{
			desc: "when one key",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"foo"},
			expected: []string{"foo"},
		},
		{
			desc: "when multiple sub keys and nested sub key",
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar"),
				aKVPair("foo/baz/2", "bar"),
				aKVPair("foo/baz/biz/1", "bar"),
			},
			keyParts: []string{"foo", "/baz/"},
			expected: []string{"foo/baz/1", "foo/baz/2"},
		},
		{
			desc:    "when KV error",
			kvError: store.ErrNotReachable,
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"foo/baz/1"},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.list(test.keyParts...)

			sort.Strings(test.expected)
			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGet(t *testing.T) {
	testCases := []struct {
		desc         string
		kvPairs      []*store.KVPair
		storeType    store.Backend
		keyParts     []string
		defaultValue string
		kvError      error
		expected     string
	}{
		{
			desc:         "when empty key parts, empty store",
			defaultValue: "circle",
			keyParts:     []string{},
			expected:     "circle",
		},
		{
			desc:         "when non existing key",
			defaultValue: "circle",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"bar"},
			expected: "circle",
		},
		{
			desc: "when one part key",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"foo"},
			expected: "bar",
		},
		{
			desc: "when several parts key",
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"foo", "/baz/", "2"},
			expected: "bar2",
		},
		{
			desc:         "when several parts key, starts with /",
			defaultValue: "circle",
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"/foo", "/baz/", "2"},
			expected: "circle",
		},
		{
			desc:      "when several parts key starts with /, ETCD v2",
			storeType: store.ETCD,
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"/foo", "/baz/", "2"},
			expected: "bar2",
		},
		{
			desc:    "when KV error",
			kvError: store.ErrNotReachable,
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"foo/baz/1"},
			expected: "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient:  newKvClientMock(test.kvPairs, test.kvError),
				storeType: test.storeType,
			}

			actual := p.get(test.defaultValue, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key %v", test.keyParts)
		})
	}
}

func TestProviderLast(t *testing.T) {
	p := &Provider{}

	testCases := []struct {
		key      string
		expected string
	}{
		{
			key:      "",
			expected: "",
		},
		{
			key:      "foo",
			expected: "foo",
		},
		{
			key:      "foo/bar",
			expected: "bar",
		},
		{
			key:      "foo/bar/baz",
			expected: "baz",
		},
		// FIXME is this wanted ?
		{
			key:      "foo/bar/",
			expected: "",
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			actual := p.last(test.key)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderSplitGet(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected []string
	}{
		{
			desc: "when has value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "courgette, carotte, tomate, aubergine"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: []string{"courgette", "carotte", "tomate", "aubergine"},
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: nil,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: nil,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrNotReachable,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			values := p.splitGet(test.keyParts...)

			assert.Equal(t, test.expected, values, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetBool(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected bool
	}{
		{
			desc: "when value is 'true",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: true,
		},
		{
			desc: "when value is 'false",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "false"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrNotReachable,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.getBool(false, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetInt(t *testing.T) {
	defaultValue := 666

	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected int
	}{
		{
			desc: "when has value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "6"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: 6,
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrNotReachable,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.getInt(defaultValue, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetInt64(t *testing.T) {
	var defaultValue int64 = 666

	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected int64
	}{
		{
			desc: "when has value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "6"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: 6,
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrNotReachable,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.getInt64(defaultValue, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderHasStickinessLabel(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		rootPath string
		expected bool
	}{
		{
			desc:     "without option",
			expected: false,
		},
		{
			desc:     "with cookie name without stickiness=true",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerStickinessCookieName, "aubergine"),
				),
			),
			expected: false,
		},
		{
			desc:     "stickiness=true",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerStickiness, "true"),
				),
			),
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: &Mock{
					KVPairs: test.kvPairs,
				},
			}

			actual := p.hasStickinessLabel(test.rootPath)

			if actual != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}
