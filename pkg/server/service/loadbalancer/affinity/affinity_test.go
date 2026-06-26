package affinity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type key string

const serviceName key = "serviceName"

func newBalancer(regex, header string) *Balancer {
	return New(&dynamic.AffinityConfig{
		Regex:      regex,
		HeaderName: header,
	}, false)
}

func addServer(b *Balancer, name string) {
	b.AddServer(name, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", name)
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})
}

func TestBalancerNoServer(t *testing.T) {
	balancer := newBalancer(`^/hello/([^/]+)`, "")

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/hello/session-1", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestPathAffinitySticky(t *testing.T) {
	balancer := newBalancer(`^/hello/([^/]+)`, "")
	addServer(balancer, "first")
	addServer(balancer, "second")
	addServer(balancer, "third")

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

	// All requests with the same session-id must land on the same backend.
	for range 10 {
		recorder.ResponseRecorder = httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/hello/abc-123", nil))
	}
	// Exactly one backend should have received all 10 requests.
	assert.True(t, recorder.save["first"] == 10 || recorder.save["second"] == 10 || recorder.save["third"] == 10)
}

func TestPathAffinityDifferentSessionsDifferentBackends(t *testing.T) {
	balancer := newBalancer(`^/hello/([^/]+)`, "")
	addServer(balancer, "first")
	addServer(balancer, "second")
	addServer(balancer, "third")

	backendFor := func(session string) string {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/hello/"+session, nil))
		return recorder.Header().Get("server")
	}

	// Sessions must be consistent across repeated calls.
	for range 5 {
		assert.Equal(t, backendFor("sess-aaa"), backendFor("sess-aaa"))
		assert.Equal(t, backendFor("sess-bbb"), backendFor("sess-bbb"))
		assert.Equal(t, backendFor("sess-ccc"), backendFor("sess-ccc"))
	}

	// Collect unique backends across many sessions, should use more than one.
	seen := map[string]bool{}
	for i := range 20 {
		seen[backendFor("session-"+string(rune('a'+i)))] = true
	}
	assert.Greater(t, len(seen), 1, "expected requests to spread across multiple backends")
}

func TestPathAffinitySeqRequestsSameBackend(t *testing.T) {
	balancer := newBalancer(`^/hello/([^/]+)`, "")
	addServer(balancer, "first")
	addServer(balancer, "second")
	addServer(balancer, "third")

	getRecorder := httptest.NewRecorder()
	balancer.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/hello/my-session", nil))
	getBackend := getRecorder.Header().Get("server")

	for i := range 5 {
		postRecorder := httptest.NewRecorder()
		balancer.ServeHTTP(postRecorder, httptest.NewRequest(http.MethodPost, "/hello/my-session/"+string(rune('0'+i)), nil))
		assert.Equal(t, getBackend, postRecorder.Header().Get("server"))
	}
}

func TestPathAffinityNoMatch(t *testing.T) {
	// Path doesn't match the regex, falls back to full path hash, still consistent.
	balancer := newBalancer(`^/hello/([^/]+)`, "")
	addServer(balancer, "first")
	addServer(balancer, "second")

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 10 {
		recorder.ResponseRecorder = httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/other/path", nil))
	}
	// No affinity key, falls back to path hash, still deterministic.
	assert.True(t, recorder.save["first"] == 10 || recorder.save["second"] == 10)
}

func TestHeaderAffinitySticky(t *testing.T) {
	balancer := newBalancer("", "X-Session-Id")
	addServer(balancer, "first")
	addServer(balancer, "second")
	addServer(balancer, "third")

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

	for range 10 {
		recorder.ResponseRecorder = httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		req.Header.Set("X-Session-Id", "user-session-xyz")
		balancer.ServeHTTP(recorder, req)
	}
	assert.True(t, recorder.save["first"] == 10 || recorder.save["second"] == 10 || recorder.save["third"] == 10)
}

func TestHeaderAffinityDifferentSessions(t *testing.T) {
	balancer := newBalancer("", "X-Session-Id")
	addServer(balancer, "first")
	addServer(balancer, "second")
	addServer(balancer, "third")

	backendFor := func(session string) string {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		req.Header.Set("X-Session-Id", session)
		balancer.ServeHTTP(recorder, req)
		return recorder.Header().Get("server")
	}

	for range 5 {
		assert.Equal(t, backendFor("sess-1"), backendFor("sess-1"))
		assert.Equal(t, backendFor("sess-2"), backendFor("sess-2"))
	}

	seen := map[string]bool{}
	for i := range 20 {
		seen[backendFor("session-"+string(rune('a'+i)))] = true
	}
	assert.Greater(t, len(seen), 1, "expected requests to spread across multiple backends")
}

func TestNilConfig(t *testing.T) {
	balancer := New(nil, false)
	addServer(balancer, "first")

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/hello/session", nil))

	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode)
}

func TestSetStatusDownEvictsSession(t *testing.T) {
	balancer := newBalancer(`^/hello/([^/]+)`, "")
	addServer(balancer, "first")
	addServer(balancer, "second")

	// Pin a session to whichever backend is chosen.
	req := httptest.NewRequest(http.MethodGet, "/hello/pinned-session", nil)
	first := httptest.NewRecorder()
	balancer.ServeHTTP(first, req)
	pinnedBackend := first.Header().Get("server")

	// Mark that backend as down, session must be evicted from the ring.
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), pinnedBackend, false)

	assert.NotContains(t, balancer.ring, "pinned-session", "evicted session should be removed from ring")
}

func TestSetStatusNoServiceUp(t *testing.T) {
	balancer := newBalancer(`^/hello/([^/]+)`, "")
	addServer(balancer, "first")
	addServer(balancer, "second")

	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "first", false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	// All servers down, balancer returns 503.
	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/hello/any", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

type responseRecorder struct {
	*httptest.ResponseRecorder
	save     map[string]int
	sequence []string
	status   []int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.save[r.Header().Get("server")]++
	r.sequence = append(r.sequence, r.Header().Get("server"))
	r.status = append(r.status, statusCode)
	r.ResponseRecorder.WriteHeader(statusCode)
}
