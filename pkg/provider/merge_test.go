package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
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
	mergeResourceMaps(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider", tracker)

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
	mergeResourceMaps(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider", tracker)

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
	mergeResourceMaps(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider", tracker)

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
	mergeResourceMaps(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider1", tracker)

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
	mergeResourceMaps(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider1", tracker)

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
	mergeResourceMaps(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), "provider1", tracker)

	// Different values, no Merge method -> conflict.
	assert.Len(t, tracker.toDelete, 1)
	assertMarkedForDeletion(t, tracker.toDelete, "res1")
}

func TestMerge_HTTPRouters(t *testing.T) {
	testCases := []struct {
		name           string
		configurations map[string]*dynamic.Configuration
		expected       map[string]*dynamic.Router
	}{
		{
			name: "multiple providers different routers",
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
			expected: map[string]*dynamic.Router{
				"router1": {Rule: "Host(`example1.com`)"},
				"router2": {Rule: "Host(`example2.com`)"},
			},
		},
		{
			name: "conflict: multiple providers same router different config",
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
			expected: map[string]*dynamic.Router{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Merge(context.Background(), tc.configurations)
			assert.Equal(t, tc.expected, result.HTTP.Routers)
		})
	}
}

func TestMerge_HTTPServices(t *testing.T) {
	testCases := []struct {
		name           string
		configurations map[string]*dynamic.Configuration
		expected       map[string]*dynamic.Service
	}{
		{
			name: "multiple providers same service, servers merged",
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
			expected: map[string]*dynamic.Service{
				"service1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{
							{URL: "http://server1:80"},
							{URL: "http://server2:80"},
						},
					},
				},
			},
		},
		{
			name: "multiple providers same service, duplicate servers deduplicated",
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
			expected: map[string]*dynamic.Service{
				"service1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{
							{URL: "http://server1:80"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Merge(context.Background(), tc.configurations)
			assert.Equal(t, tc.expected, result.HTTP.Services)
		})
	}
}

func TestMerge_NilConfigurations(t *testing.T) {
	configurations := map[string]*dynamic.Configuration{
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
	}

	result := Merge(context.Background(), configurations)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTP)
	assert.Len(t, result.HTTP.Routers, 1)
}

func TestMerge_EmptyConfigurations(t *testing.T) {
	configurations := map[string]*dynamic.Configuration{}

	result := Merge(context.Background(), configurations)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTP)
	require.NotNil(t, result.TCP)
	require.NotNil(t, result.UDP)
	require.NotNil(t, result.TLS)
}

// assertMarkedForDeletion checks that toDelete contains an entry with the given key.
func assertMarkedForDeletion(t *testing.T, toDelete map[conflictKey]conflictInfo, key string) {
	t.Helper()
	for ck := range toDelete {
		if ck.key == key {
			return
		}
	}
	t.Errorf("toDelete does not contain key %q", key)
}
