package kv

import (
	"errors"
	"reflect"
	"sort"
	"strings"
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
				Kvclient: &Mock{},
			},
			keys:     []string{},
			expected: []string{},
		},
		{
			provider: &Provider{
				Kvclient: &Mock{},
			},
			keys:     []string{"traefik"},
			expected: []string{},
		},
		{
			provider: &Provider{
				Kvclient: &Mock{
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
				Kvclient: &Mock{
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
				Kvclient: &Mock{
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
		Kvclient: &Mock{
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
				Kvclient: &Mock{},
			},
			keys:     []string{},
			expected: "",
		},
		{
			provider: &Provider{
				Kvclient: &Mock{},
			},
			keys:     []string{"traefik"},
			expected: "",
		},
		{
			provider: &Provider{
				Kvclient: &Mock{
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
				Kvclient: &Mock{
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
				Kvclient: &Mock{
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
		Kvclient: &Mock{
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

type KvMock struct {
	Provider
}

func (provider *KvMock) loadConfig() *types.Configuration {
	return nil
}

func TestKvWatchTree(t *testing.T) {
	returnedChans := make(chan chan []*store.KVPair)
	provider := &KvMock{
		Provider{
			Kvclient: &Mock{
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
	case _ = <-configChan:
		t.Fatalf("configChan should be empty")
	default:
	}
}

// Override Get/List to return a error
type KvError struct {
	Get  error
	List error
}

// Extremely limited mock store so we can test initialization
type Mock struct {
	Error           KvError
	KVPairs         []*store.KVPair
	WatchTreeMethod func() <-chan []*store.KVPair
}

func (s *Mock) Put(key string, value []byte, opts *store.WriteOptions) error {
	return errors.New("Put not supported")
}

func (s *Mock) Get(key string) (*store.KVPair, error) {
	if err := s.Error.Get; err != nil {
		return nil, err
	}
	for _, kvPair := range s.KVPairs {
		if kvPair.Key == key {
			return kvPair, nil
		}
	}
	return nil, store.ErrKeyNotFound
}

func (s *Mock) Delete(key string) error {
	return errors.New("Delete not supported")
}

// Exists mock
func (s *Mock) Exists(key string) (bool, error) {
	return false, errors.New("Exists not supported")
}

// Watch mock
func (s *Mock) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return nil, errors.New("Watch not supported")
}

// WatchTree mock
func (s *Mock) WatchTree(prefix string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return s.WatchTreeMethod(), nil
}

// NewLock mock
func (s *Mock) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return nil, errors.New("NewLock not supported")
}

// List mock
func (s *Mock) List(prefix string) ([]*store.KVPair, error) {
	if err := s.Error.List; err != nil {
		return nil, err
	}
	kv := []*store.KVPair{}
	for _, kvPair := range s.KVPairs {
		if strings.HasPrefix(kvPair.Key, prefix) && !strings.ContainsAny(strings.TrimPrefix(kvPair.Key, prefix), "/") {
			kv = append(kv, kvPair)
		}
	}
	return kv, nil
}

// DeleteTree mock
func (s *Mock) DeleteTree(prefix string) error {
	return errors.New("DeleteTree not supported")
}

// AtomicPut mock
func (s *Mock) AtomicPut(key string, value []byte, previous *store.KVPair, opts *store.WriteOptions) (bool, *store.KVPair, error) {
	return false, nil, errors.New("AtomicPut not supported")
}

// AtomicDelete mock
func (s *Mock) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return false, errors.New("AtomicDelete not supported")
}

// Close mock
func (s *Mock) Close() {
	return
}

func TestKVLoadConfig(t *testing.T) {
	provider := &Provider{
		Prefix: "traefik",
		Kvclient: &Mock{
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
