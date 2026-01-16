package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tls"
)

// testResource is a simple type without a Merge method.
type testResource struct {
	Value string
}

// testMergeableResource implements the Merge method.
type testMergeableResource struct {
	Config  string
	Servers []string
}

// Merge merges another testMergeableResource into this one.
// Returns true if the merge succeeds, false if configurations conflict.
func (r *testMergeableResource) Merge(other *testMergeableResource) bool {
	if r.Config != other.Config {
		return false
	}
	r.Servers = append(r.Servers, other.Servers...)

	return true
}

// testCollectionSet is a container with a map field.
type testCollectionSet struct {
	Resources map[string]*testResource
}

// testMergeableCollectionSet is a container with a mergeable map field.
type testMergeableCollectionSet struct {
	Resources map[string]*testMergeableResource
}

// TestEmbedded is an embedded struct for testing anonymous field handling.
// Must be exported for reflection to process it.
type TestEmbedded struct {
	EmbeddedItems map[string]*testResource
}

// testCollectionSetWithEmbedded has both embedded and direct map fields.
type testCollectionSetWithEmbedded struct {
	TestEmbedded

	Items map[string]*testResource
}

func TestMergeCollectionSet_BasicMapMerge(t *testing.T) {
	dst := &testCollectionSet{
		Resources: map[string]*testResource{
			"existing": {Value: "dst"},
		},
	}
	src := &testCollectionSet{
		Resources: map[string]*testResource{
			"new": {Value: "src"},
		},
	}

	tracker := newMergeTracker()
	mergeResourceMaps(context.Background(), reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider", tracker, ResourceStrategyMerge, nil)

	assert.Empty(t, tracker.toDelete)
	assert.Equal(t, &testCollectionSet{
		Resources: map[string]*testResource{
			"existing": {Value: "dst"},
			"new":      {Value: "src"},
		},
	}, dst)
}

func TestMergeCollectionSet_EmbeddedStruct(t *testing.T) {
	dst := &testCollectionSetWithEmbedded{
		TestEmbedded: TestEmbedded{
			EmbeddedItems: map[string]*testResource{
				"embedded1": {Value: "dst-embedded"},
			},
		},
		Items: map[string]*testResource{
			"item1": {Value: "dst-item"},
		},
	}
	src := &testCollectionSetWithEmbedded{
		TestEmbedded: TestEmbedded{
			EmbeddedItems: map[string]*testResource{
				"embedded2": {Value: "src-embedded"},
			},
		},
		Items: map[string]*testResource{
			"item2": {Value: "src-item"},
		},
	}

	tracker := newMergeTracker()
	mergeResourceMaps(context.Background(), reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider", tracker, ResourceStrategyMerge, nil)

	assert.Empty(t, tracker.toDelete)
	assert.Equal(t, &testCollectionSetWithEmbedded{
		TestEmbedded: TestEmbedded{
			EmbeddedItems: map[string]*testResource{
				"embedded1": {Value: "dst-embedded"},
				"embedded2": {Value: "src-embedded"},
			},
		},
		Items: map[string]*testResource{
			"item1": {Value: "dst-item"},
			"item2": {Value: "src-item"},
		},
	}, dst)
}

func TestMergeCollectionSet_MergeableInterface(t *testing.T) {
	dst := &testMergeableCollectionSet{
		Resources: map[string]*testMergeableResource{
			"svc1": {Config: "same", Servers: []string{"server1"}},
		},
	}
	src := &testMergeableCollectionSet{
		Resources: map[string]*testMergeableResource{
			"svc1": {Config: "same", Servers: []string{"server2"}},
		},
	}

	tracker := newMergeTracker()
	mergeResourceMaps(context.Background(), reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider", tracker, ResourceStrategyMerge, nil)

	assert.Empty(t, tracker.toDelete)
	assert.Equal(t, &testMergeableCollectionSet{
		Resources: map[string]*testMergeableResource{
			"svc1": {Config: "same", Servers: []string{"server1", "server2"}},
		},
	}, dst)
}

func TestMergeCollectionSet_MergeableConflict(t *testing.T) {
	dst := &testMergeableCollectionSet{
		Resources: map[string]*testMergeableResource{
			"svc1": {Config: "config-A", Servers: []string{"server1"}},
		},
	}
	src := &testMergeableCollectionSet{
		Resources: map[string]*testMergeableResource{
			"svc1": {Config: "config-B", Servers: []string{"server2"}},
		},
	}

	tracker := newMergeTracker()
	mergeResourceMaps(context.Background(), reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider1", tracker, ResourceStrategyMerge, nil)

	// Merge() returns false due to config mismatch -> marked for deletion.
	assert.Len(t, tracker.toDelete, 1)
	assertMarkedForDeletion(t, tracker.toDelete, "svc1")
}

func TestMergeCollectionSet_DeepEqualFallback(t *testing.T) {
	dst := &testCollectionSet{
		Resources: map[string]*testResource{
			"res1": {Value: "same"},
		},
	}
	src := &testCollectionSet{
		Resources: map[string]*testResource{
			"res1": {Value: "same"},
		},
	}

	tracker := newMergeTracker()
	mergeResourceMaps(context.Background(), reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider1", tracker, ResourceStrategyMerge, nil)

	// Same values -> no conflict.
	assert.Empty(t, tracker.toDelete)
	assert.Equal(t, &testCollectionSet{
		Resources: map[string]*testResource{
			"res1": {Value: "same"},
		},
	}, dst)
}

func TestMergeCollectionSet_DeepEqualConflict(t *testing.T) {
	dst := &testCollectionSet{
		Resources: map[string]*testResource{
			"res1": {Value: "value-A"},
		},
	}
	src := &testCollectionSet{
		Resources: map[string]*testResource{
			"res1": {Value: "value-B"},
		},
	}

	tracker := newMergeTracker()
	mergeResourceMaps(context.Background(), reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider1", tracker, ResourceStrategyMerge, nil)

	// Different values, no Merge method -> conflict.
	assert.Len(t, tracker.toDelete, 1)
	assertMarkedForDeletion(t, tracker.toDelete, "res1")
}

func TestMerge(t *testing.T) {
	testCases := []struct {
		desc           string
		configurations map[string]*dynamic.Configuration
		strategy       ResourceStrategy
		expected       *dynamic.Configuration
	}{
		{
			desc: "HTTP routers: multiple providers different routers",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router1": {Rule: "Host(`example1.com`)"},
						},
					},
				},
				"provider2": {
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router2": {Rule: "Host(`example2.com`)"},
						},
					},
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.HTTP.Routers["router1"] = &dynamic.Router{Rule: "Host(`example1.com`)"}
				c.HTTP.Routers["router2"] = &dynamic.Router{Rule: "Host(`example2.com`)"}
			}),
		},
		{
			desc: "HTTP routers: conflict multiple providers same router different config",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router1": {Rule: "Host(`example1.com`)"},
						},
					},
				},
				"provider2": {
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router1": {Rule: "Host(`example2.com`)"},
						},
					},
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(nil),
		},
		{
			desc: "HTTP services: multiple providers same service servers merged",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					HTTP: &dynamic.HTTPConfiguration{
						Services: map[string]*dynamic.Service{
							"service1": {
								LoadBalancer: &dynamic.ServersLoadBalancer{
									Servers: []dynamic.Server{
										{URL: "http://server1:80"},
									},
								},
							},
						},
					},
				},
				"provider2": {
					HTTP: &dynamic.HTTPConfiguration{
						Services: map[string]*dynamic.Service{
							"service1": {
								LoadBalancer: &dynamic.ServersLoadBalancer{
									Servers: []dynamic.Server{
										{URL: "http://server2:80"},
									},
								},
							},
						},
					},
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.HTTP.Services["service1"] = &dynamic.Service{
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{
							{URL: "http://server1:80"},
							{URL: "http://server2:80"},
						},
					},
				}
			}),
		},
		{
			desc: "HTTP services: multiple providers same service duplicate servers deduplicated",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					HTTP: &dynamic.HTTPConfiguration{
						Services: map[string]*dynamic.Service{
							"service1": {
								LoadBalancer: &dynamic.ServersLoadBalancer{
									Servers: []dynamic.Server{
										{URL: "http://server1:80"},
									},
								},
							},
						},
					},
				},
				"provider2": {
					HTTP: &dynamic.HTTPConfiguration{
						Services: map[string]*dynamic.Service{
							"service1": {
								LoadBalancer: &dynamic.ServersLoadBalancer{
									Servers: []dynamic.Server{
										{URL: "http://server1:80"},
									},
								},
							},
						},
					},
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.HTTP.Services["service1"] = &dynamic.Service{
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{
							{URL: "http://server1:80"},
						},
					},
				}
			}),
		},
		{
			desc: "TLS certificates: different certificates both kept",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert1.pem", KeyFile: "key1.pem"},
								Stores:      []string{"store1"},
							},
						},
					},
				},
				"provider2": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert2.pem", KeyFile: "key2.pem"},
								Stores:      []string{"store2"},
							},
						},
					},
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.TLS.Certificates = []*tls.CertAndStores{
					{
						Certificate: tls.Certificate{CertFile: "cert1.pem", KeyFile: "key1.pem"},
						Stores:      []string{"store1"},
					},
					{
						Certificate: tls.Certificate{CertFile: "cert2.pem", KeyFile: "key2.pem"},
						Stores:      []string{"store2"},
					},
				}
			}),
		},
		{
			desc: "TLS certificates: same certificate stores merged with ResourceStrategyMerge",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
								Stores:      []string{"store1"},
							},
						},
					},
				},
				"provider2": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
								Stores:      []string{"store2"},
							},
						},
					},
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.TLS.Certificates = []*tls.CertAndStores{
					{
						Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
						Stores:      []string{"store1", "store2"},
					},
				}
			}),
		},
		{
			desc: "TLS certificates: same certificate overlapping stores deduplicated",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
								Stores:      []string{"store1", "store2"},
							},
						},
					},
				},
				"provider2": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
								Stores:      []string{"store2", "store3"},
							},
						},
					},
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.TLS.Certificates = []*tls.CertAndStores{
					{
						Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
						Stores:      []string{"store1", "store2", "store3"},
					},
				}
			}),
		},
		{
			desc: "TLS certificates: same certificate stores not merged with ResourceStrategySkipDuplicates",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
								Stores:      []string{"store1"},
							},
						},
					},
				},
				"provider2": {
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{
								Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
								Stores:      []string{"store2"},
							},
						},
					},
				},
			},
			strategy: ResourceStrategySkipDuplicates,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.TLS.Certificates = []*tls.CertAndStores{
					{
						Certificate: tls.Certificate{CertFile: "cert.pem", KeyFile: "key.pem"},
						Stores:      []string{"store1"},
					},
				}
			}),
		},
		{
			desc: "nil configuration from one provider",
			configurations: map[string]*dynamic.Configuration{
				"provider1": {
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router1": {Rule: "Host(`example.com`)"},
						},
					},
				},
				"provider2": {
					// No HTTP configuration
				},
			},
			strategy: ResourceStrategyMerge,
			expected: buildExpectedConfiguration(func(c *dynamic.Configuration) {
				c.HTTP.Routers["router1"] = &dynamic.Router{Rule: "Host(`example.com`)"}
			}),
		},
		{
			desc:           "empty configurations",
			configurations: map[string]*dynamic.Configuration{},
			strategy:       ResourceStrategyMerge,
			expected:       buildExpectedConfiguration(nil),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			result := Merge(context.Background(), test.configurations, test.strategy)

			assert.Equal(t, test.expected, result)
		})
	}
}

func buildExpectedConfiguration(modifier func(*dynamic.Configuration)) *dynamic.Configuration {
	c := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			Middlewares:       make(map[string]*dynamic.TCPMiddleware),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores: make(map[string]tls.Store),
		},
	}
	if modifier != nil {
		modifier(c)
	}
	return c
}

// assertMarkedForDeletion checks that toDelete contains an entry with the given key.
func assertMarkedForDeletion(t *testing.T, toDelete map[conflictKey]conflictInfo, key string) {
	t.Helper()
	for ck := range toDelete {
		if ck.resourceKey == key {
			return
		}
	}
	t.Errorf("toDelete does not contain key %q", key)
}
