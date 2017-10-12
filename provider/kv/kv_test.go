package kv

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
)

func TestKvList(t *testing.T) {
	cases := []struct {
		provider *Provider
		keys     []string
		expected []string
	}{
		{
			provider: &Provider{
				kvclient: &Mock{},
			},
			keys:     []string{},
			expected: []string{},
		},
		{
			provider: &Provider{
				kvclient: &Mock{},
			},
			keys:     []string{"traefik"},
			expected: []string{},
		},
		{
			provider: &Provider{
				kvclient: &Mock{
					KVPairs: []*store.KVPair{
						{
							Key:   "foo",
							Value: []byte("bar"),
						},
					},
				},
			},
			keys:     []string{"bar"},
			expected: []string{},
		},
		{
			provider: &Provider{
				kvclient: &Mock{
					KVPairs: []*store.KVPair{
						{
							Key:   "foo",
							Value: []byte("bar"),
						},
					},
				},
			},
			keys:     []string{"foo"},
			expected: []string{"foo"},
		},
		{
			provider: &Provider{
				kvclient: &Mock{
					KVPairs: []*store.KVPair{
						{
							Key:   "foo/baz/1",
							Value: []byte("bar"),
						},
						{
							Key:   "foo/baz/2",
							Value: []byte("bar"),
						},
						{
							Key:   "foo/baz/biz/1",
							Value: []byte("bar"),
						},
					},
				},
			},
			keys:     []string{"foo", "/baz/"},
			expected: []string{"foo/baz/1", "foo/baz/2"},
		},
	}

	for _, c := range cases {
		actual := c.provider.list(c.keys...)
		sort.Strings(actual)
		sort.Strings(c.expected)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected %v, got %v for %v and %v", c.expected, actual, c.keys, c.provider)
		}
	}

	// Error case
	provider := &Provider{
		kvclient: &Mock{
			Error: KvError{
				List: store.ErrKeyNotFound,
			},
		},
	}
	actual := provider.list("anything")
	if actual != nil {
		t.Fatalf("Should have return nil, got %v", actual)
	}
}

func TestKvGet(t *testing.T) {
	cases := []struct {
		provider *Provider
		keys     []string
		expected string
	}{
		{
			provider: &Provider{
				kvclient: &Mock{},
			},
			keys:     []string{},
			expected: "",
		},
		{
			provider: &Provider{
				kvclient: &Mock{},
			},
			keys:     []string{"traefik"},
			expected: "",
		},
		{
			provider: &Provider{
				kvclient: &Mock{
					KVPairs: []*store.KVPair{
						{
							Key:   "foo",
							Value: []byte("bar"),
						},
					},
				},
			},
			keys:     []string{"bar"},
			expected: "",
		},
		{
			provider: &Provider{
				kvclient: &Mock{
					KVPairs: []*store.KVPair{
						{
							Key:   "foo",
							Value: []byte("bar"),
						},
					},
				},
			},
			keys:     []string{"foo"},
			expected: "bar",
		},
		{
			provider: &Provider{
				kvclient: &Mock{
					KVPairs: []*store.KVPair{
						{
							Key:   "foo/baz/1",
							Value: []byte("bar1"),
						},
						{
							Key:   "foo/baz/2",
							Value: []byte("bar2"),
						},
						{
							Key:   "foo/baz/biz/1",
							Value: []byte("bar3"),
						},
					},
				},
			},
			keys:     []string{"foo", "/baz/", "2"},
			expected: "bar2",
		},
	}

	for _, c := range cases {
		actual := c.provider.get("", c.keys...)
		if actual != c.expected {
			t.Fatalf("expected %v, got %v for %v and %v", c.expected, actual, c.keys, c.provider)
		}
	}

	// Error case
	provider := &Provider{
		kvclient: &Mock{
			Error: KvError{
				Get: store.ErrKeyNotFound,
			},
		},
	}
	actual := provider.get("", "anything")
	if actual != "" {
		t.Fatalf("Should have return nil, got %v", actual)
	}
}

func TestKvLast(t *testing.T) {
	cases := []struct {
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

	provider := &Provider{}
	for _, c := range cases {
		actual := provider.last(c.key)
		if actual != c.expected {
			t.Fatalf("expected %s, got %s", c.expected, actual)
		}
	}
}

func TestKvWatchTree(t *testing.T) {
	returnedChans := make(chan chan []*store.KVPair)
	provider := &KvMock{
		Provider{
			kvclient: &Mock{
				WatchTreeMethod: func() <-chan []*store.KVPair {
					c := make(chan []*store.KVPair, 10)
					returnedChans <- c
					return c
				},
			},
		},
	}

	configChan := make(chan types.ConfigMessage)
	go func() {
		provider.watchKv(configChan, "prefix", make(chan bool, 1))
	}()

	select {
	case c1 := <-returnedChans:
		c1 <- []*store.KVPair{}
		<-configChan
		close(c1) // WatchTree chans can close due to error
	case <-time.After(1 * time.Second):
		t.Fatalf("Failed to create a new WatchTree chan")
	}

	select {
	case c2 := <-returnedChans:
		c2 <- []*store.KVPair{}
		<-configChan
	case <-time.After(1 * time.Second):
		t.Fatalf("Failed to create a new WatchTree chan")
	}

	select {
	case <-configChan:
		t.Fatalf("configChan should be empty")
	default:
	}
}

func TestKVLoadConfig(t *testing.T) {
	provider := &Provider{
		Prefix: "traefik",
		kvclient: &Mock{
			KVPairs: []*store.KVPair{
				{
					Key:   "traefik/frontends/frontend.with.dot",
					Value: []byte(""),
				},
				{
					Key:   "traefik/frontends/frontend.with.dot/backend",
					Value: []byte("backend.with.dot.too"),
				},
				{
					Key:   "traefik/frontends/frontend.with.dot/routes",
					Value: []byte(""),
				},
				{
					Key:   "traefik/frontends/frontend.with.dot/routes/route.with.dot",
					Value: []byte(""),
				},
				{
					Key:   "traefik/frontends/frontend.with.dot/routes/route.with.dot/rule",
					Value: []byte("Host:test.localhost"),
				},
				{
					Key:   "traefik/backends/backend.with.dot.too",
					Value: []byte(""),
				},
				{
					Key:   "traefik/backends/backend.with.dot.too/servers",
					Value: []byte(""),
				},
				{
					Key:   "traefik/backends/backend.with.dot.too/servers/server.with.dot",
					Value: []byte(""),
				},
				{
					Key:   "traefik/backends/backend.with.dot.too/servers/server.with.dot/url",
					Value: []byte("http://172.17.0.2:80"),
				},
				{
					Key:   "traefik/backends/backend.with.dot.too/servers/server.with.dot/weight",
					Value: []byte("0"),
				},
				{
					Key:   "traefik/backends/backend.with.dot.too/servers/server.with.dot.without.url",
					Value: []byte(""),
				},
				{
					Key:   "traefik/backends/backend.with.dot.too/servers/server.with.dot.without.url/weight",
					Value: []byte("0"),
				},
			},
		},
	}
	actual := provider.loadConfig()
	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"backend.with.dot.too": {
				Servers: map[string]types.Server{
					"server.with.dot": {
						URL:    "http://172.17.0.2:80",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
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
	}
	if !reflect.DeepEqual(actual.Backends, expected.Backends) {
		t.Fatalf("expected %+v, got %+v", expected.Backends, actual.Backends)
	}
	if !reflect.DeepEqual(actual.Frontends, expected.Frontends) {
		t.Fatalf("expected %+v, got %+v", expected.Frontends, actual.Frontends)
	}
}

func TestKVHasStickinessLabel(t *testing.T) {
	testCases := []struct {
		desc     string
		KVPairs  []*store.KVPair
		expected bool
	}{
		{
			desc:     "without option",
			expected: false,
		},
		{
			desc: "with cookie name without stickiness=true",
			KVPairs: []*store.KVPair{
				{
					Key:   "loadbalancer/stickiness/cookiename",
					Value: []byte("foo"),
				},
			},
			expected: false,
		},
		{
			desc: "stickiness=true",
			KVPairs: []*store.KVPair{
				{
					Key:   "loadbalancer/stickiness",
					Value: []byte("true"),
				},
			},
			expected: true,
		},
		{
			desc: "stickiness=true and sticky=true",
			KVPairs: []*store.KVPair{
				{
					Key:   "loadbalancer/stickiness",
					Value: []byte("true"),
				},
				{
					Key:   "loadbalancer/sticky",
					Value: []byte("true"),
				},
			},
			expected: true,
		},
		{
			desc: "stickiness=false and sticky=true",
			KVPairs: []*store.KVPair{
				{
					Key:   "loadbalancer/stickiness",
					Value: []byte("false"),
				},
				{
					Key:   "loadbalancer/sticky",
					Value: []byte("true"),
				},
			},
			expected: true,
		},
		{
			desc: "stickiness=true and sticky=false",
			KVPairs: []*store.KVPair{
				{
					Key:   "loadbalancer/stickiness",
					Value: []byte("true"),
				},
				{
					Key:   "loadbalancer/sticky",
					Value: []byte("false"),
				},
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvclient: &Mock{
					KVPairs: test.KVPairs,
				},
			}

			actual := p.hasStickinessLabel("")

			if actual != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}
