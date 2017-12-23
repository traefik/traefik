package kv

import (
	"sort"
	"strings"
	"testing"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
)

type ByKey []*store.KVPair

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

func filler(prefix string, opts ...func(string, map[string]*store.KVPair)) []*store.KVPair {
	buf := make(map[string]*store.KVPair)
	for _, opt := range opts {
		opt(prefix, buf)
	}

	var result ByKey
	for _, value := range buf {
		result = append(result, value)
	}

	sort.Sort(result)
	return result
}

func backend(name string, opts ...func(map[string]string)) func(string, map[string]*store.KVPair) {
	return entry(pathBackends+name, opts...)
}

func frontend(name string, opts ...func(map[string]string)) func(string, map[string]*store.KVPair) {
	return entry(pathFrontends+name, opts...)
}

func entry(root string, opts ...func(map[string]string)) func(string, map[string]*store.KVPair) {
	return func(prefix string, pairs map[string]*store.KVPair) {
		prefixedRoot := prefix + pathSeparator + strings.TrimPrefix(root, pathSeparator)
		pairs[prefixedRoot] = &store.KVPair{Key: prefixedRoot, Value: []byte("")}

		transit := make(map[string]string)
		for _, opt := range opts {
			opt(transit)
		}

		for key, value := range transit {
			fill(pairs, prefixedRoot, key, value)
		}
	}
}

func fill(pairs map[string]*store.KVPair, previous string, current string, value string) {
	clean := strings.TrimPrefix(current, pathSeparator)

	i := strings.IndexRune(clean, '/')
	if i > 0 {
		key := previous + pathSeparator + clean[:i]

		if _, ok := pairs[key]; !ok || len(pairs[key].Value) == 0 {
			pairs[key] = &store.KVPair{Key: key, Value: []byte("")}
		}

		fill(pairs, key, clean[i:], value)
	}

	key := previous + pathSeparator + clean
	pairs[key] = &store.KVPair{Key: key, Value: []byte(value)}
}

func withPair(key string, value string) func(map[string]string) {
	return func(pairs map[string]string) {
		if len(key) == 0 {
			return
		}
		pairs[key] = value
	}
}

func TestFiller(t *testing.T) {
	expected := []*store.KVPair{
		{Key: "traefik/backends/backend.with.dot.too", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot.without.url", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot.without.url/weight", Value: []byte("0")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot/url", Value: []byte("http://172.17.0.2:80")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot/weight", Value: []byte("0")},
		{Key: "traefik/frontends/frontend.with.dot", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/backend", Value: []byte("backend.with.dot.too")},
		{Key: "traefik/frontends/frontend.with.dot/routes", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/routes/route.with.dot", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/routes/route.with.dot/rule", Value: []byte("Host:test.localhost")},
	}

	pairs1 := filler("traefik",
		frontend("frontend.with.dot",
			withPair("backend", "backend.with.dot.too"),
			withPair("routes/route.with.dot/rule", "Host:test.localhost")),
		backend("backend.with.dot.too",
			withPair("servers/server.with.dot/url", "http://172.17.0.2:80"),
			withPair("servers/server.with.dot/weight", "0"),
			withPair("servers/server.with.dot.without.url/weight", "0")),
	)
	assert.EqualValues(t, expected, pairs1)

	pairs2 := filler("traefik",
		entry("frontends/frontend.with.dot",
			withPair("backend", "backend.with.dot.too"),
			withPair("routes/route.with.dot/rule", "Host:test.localhost")),
		entry("backends/backend.with.dot.too",
			withPair("servers/server.with.dot/url", "http://172.17.0.2:80"),
			withPair("servers/server.with.dot/weight", "0"),
			withPair("servers/server.with.dot.without.url/weight", "0")),
	)
	assert.EqualValues(t, expected, pairs2)
}
