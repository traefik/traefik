package kv

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/abronan/valkeyrie/store"
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

func withList(key string, values ...string) func(map[string]string) {
	return func(pairs map[string]string) {
		if len(key) == 0 {
			return
		}
		for i, value := range values {
			pairs[key+"/"+strconv.Itoa(i)] = value
		}
	}
}

func withErrorPage(name string, backend string, query string, statuses ...string) func(map[string]string) {
	return func(pairs map[string]string) {
		if len(name) == 0 {
			return
		}

		withPair(pathFrontendErrorPages+name+pathFrontendErrorPagesBackend, backend)(pairs)
		withPair(pathFrontendErrorPages+name+pathFrontendErrorPagesQuery, query)(pairs)
		withList(pathFrontendErrorPages+name+pathFrontendErrorPagesStatus, statuses...)(pairs)
	}
}

func withRateLimit(extractorFunc string, opts ...func(map[string]string)) func(map[string]string) {
	return func(pairs map[string]string) {
		pairs[pathFrontendRateLimitExtractorFunc] = extractorFunc
		for _, opt := range opts {
			opt(pairs)
		}
	}
}

func withLimit(name string, average, burst, period string) func(map[string]string) {
	return func(pairs map[string]string) {
		pairs[pathFrontendRateLimitRateSet+name+pathFrontendRateLimitAverage] = average
		pairs[pathFrontendRateLimitRateSet+name+pathFrontendRateLimitBurst] = burst
		pairs[pathFrontendRateLimitRateSet+name+pathFrontendRateLimitPeriod] = period
	}
}

func TestFiller(t *testing.T) {
	expected := []*store.KVPair{
		{Key: "traefik/backends/backend.with.dot.too", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot.without.url", Value: []byte("")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot.without.url/weight", Value: []byte("1")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot/url", Value: []byte("http://172.17.0.2:80")},
		{Key: "traefik/backends/backend.with.dot.too/servers/server.with.dot/weight", Value: []byte("1")},
		{Key: "traefik/frontends/frontend.with.dot", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/backend", Value: []byte("backend.with.dot.too")},
		{Key: "traefik/frontends/frontend.with.dot/errors", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/errors/bar", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/errors/bar/backend", Value: []byte("error")},
		{Key: "traefik/frontends/frontend.with.dot/errors/bar/query", Value: []byte("/test2")},
		{Key: "traefik/frontends/frontend.with.dot/errors/bar/status", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/errors/bar/status/0", Value: []byte("400-405")},
		{Key: "traefik/frontends/frontend.with.dot/errors/foo", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/errors/foo/backend", Value: []byte("error")},
		{Key: "traefik/frontends/frontend.with.dot/errors/foo/query", Value: []byte("/test1")},
		{Key: "traefik/frontends/frontend.with.dot/errors/foo/status", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/errors/foo/status/0", Value: []byte("500-501")},
		{Key: "traefik/frontends/frontend.with.dot/errors/foo/status/1", Value: []byte("503-599")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/extractorfunc", Value: []byte("client.ip")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/bar", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/bar/average", Value: []byte("3")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/bar/burst", Value: []byte("6")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/bar/period", Value: []byte("9")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/foo", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/foo/average", Value: []byte("6")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/foo/burst", Value: []byte("12")},
		{Key: "traefik/frontends/frontend.with.dot/ratelimit/rateset/foo/period", Value: []byte("18")},
		{Key: "traefik/frontends/frontend.with.dot/routes", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/routes/route.with.dot", Value: []byte("")},
		{Key: "traefik/frontends/frontend.with.dot/routes/route.with.dot/rule", Value: []byte("Host:test.localhost")},
	}

	pairs1 := filler("traefik",
		frontend("frontend.with.dot",
			withPair("backend", "backend.with.dot.too"),
			withPair("routes/route.with.dot/rule", "Host:test.localhost"),
			withErrorPage("foo", "error", "/test1", "500-501", "503-599"),
			withErrorPage("bar", "error", "/test2", "400-405"),
			withRateLimit("client.ip",
				withLimit("foo", "6", "12", "18"),
				withLimit("bar", "3", "6", "9"))),
		backend("backend.with.dot.too",
			withPair("servers/server.with.dot/url", "http://172.17.0.2:80"),
			withPair("servers/server.with.dot/weight", "1"),
			withPair("servers/server.with.dot.without.url/weight", "1")),
	)
	assert.EqualValues(t, expected, pairs1)

	pairs2 := filler("traefik",
		entry("frontends/frontend.with.dot",
			withPair("backend", "backend.with.dot.too"),
			withPair("routes/route.with.dot/rule", "Host:test.localhost"),
			withPair("errors/foo/backend", "error"),
			withPair("errors/foo/query", "/test1"),
			withList("errors/foo/status", "500-501", "503-599"),
			withPair("errors/bar/backend", "error"),
			withPair("errors/bar/query", "/test2"),
			withList("errors/bar/status", "400-405"),
			withPair("ratelimit/extractorfunc", "client.ip"),
			withPair("ratelimit/rateset/foo/average", "6"),
			withPair("ratelimit/rateset/foo/burst", "12"),
			withPair("ratelimit/rateset/foo/period", "18"),
			withPair("ratelimit/rateset/bar/average", "3"),
			withPair("ratelimit/rateset/bar/burst", "6"),
			withPair("ratelimit/rateset/bar/period", "9")),
		entry("backends/backend.with.dot.too",
			withPair("servers/server.with.dot/url", "http://172.17.0.2:80"),
			withPair("servers/server.with.dot/weight", "1"),
			withPair("servers/server.with.dot.without.url/weight", "1")),
	)
	assert.EqualValues(t, expected, pairs2)
}
