package p2c

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer"
)

func TestP2C(t *testing.T) {
	testCases := []struct {
		name            string
		handlers        []*namedHandler
		rand            *mockRand
		expectedHandler string
	}{
		{
			name:            "oneHealthyHandler",
			handlers:        testHandlers(0),
			rand:            nil,
			expectedHandler: "0",
		},
		{
			name:            "twoHandlersZeroInflight",
			handlers:        testHandlers(0, 0),
			rand:            &mockRand{vals: []int{1, 0}},
			expectedHandler: "1",
		},
		{
			name:            "choosesLowerOfTwo",
			handlers:        testHandlers(0, 1),
			rand:            &mockRand{vals: []int{1, 0}},
			expectedHandler: "0",
		},
		{
			name:            "choosesLowerOfThree",
			handlers:        testHandlers(10, 90, 40),
			rand:            &mockRand{vals: []int{1, 1}},
			expectedHandler: "2",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			strategy := New(nil, false)
			strategy.rand = test.rand

			for _, h := range test.handlers {
				strategy.handlers = append(strategy.handlers, h)
				strategy.status[h.name] = struct{}{}
			}

			got, err := strategy.nextServer()
			require.NoError(t, err)

			assert.Equal(t, test.expectedHandler, got.name, "balancer strategy gave unexpected backend handler")
		})
	}
}

func TestSticky(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{
			Name:     "test",
			Secure:   true,
			HTTPOnly: true,
			SameSite: "none",
			MaxAge:   42,
			Path:     func(v string) *string { return &v }("/foo"),
		},
	}, false)
	balancer.rand = &mockRand{vals: []int{1, 0}}

	balancer.AddServer("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})

	balancer.AddServer("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})

	recorder := &responseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		save:             map[string]int{},
		cookies:          make(map[string]*http.Cookie),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for range 3 {
		for _, cookie := range recorder.Result().Cookies() {
			assert.NotContains(t, "first", cookie.Value)
			assert.NotContains(t, "second", cookie.Value)
			req.AddCookie(cookie)
		}
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, req)
	}

	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 3, recorder.save["second"])
	assert.True(t, recorder.cookies["test"].HttpOnly)
	assert.True(t, recorder.cookies["test"].Secure)
	assert.Equal(t, http.SameSiteNoneMode, recorder.cookies["test"].SameSite)
	assert.Equal(t, 42, recorder.cookies["test"].MaxAge)
	assert.Equal(t, "/foo", recorder.cookies["test"].Path)
}

func TestSticky_FallBack(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{Name: "test"},
	}, false)
	balancer.rand = &mockRand{vals: []int{1, 0}}

	balancer.AddServer("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})

	balancer.AddServer("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}, cookies: make(map[string]*http.Cookie)}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: "second"})
	for range 3 {
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, req)
	}

	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 3, recorder.save["second"])
}

func TestBalancerPropagate(t *testing.T) {
	balancer := New(nil, true)

	balancer.AddServer("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})
	balancer.AddServer("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})

	var calls int
	err := balancer.RegisterStatusUpdater(func(up bool) {
		calls++
	})
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)

	// two gets downed, but balancer still up since first is still up.
	balancer.SetStatus(context.Background(), "second", false)
	assert.Equal(t, 0, calls)

	recorder = httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "first", recorder.Header().Get("server"))

	// first gets downed, balancer is down.
	balancer.SetStatus(context.Background(), "first", false)
	assert.Equal(t, 1, calls)

	recorder = httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)

	// two gets up, balancer up.
	balancer.SetStatus(context.Background(), "second", true)
	assert.Equal(t, 2, calls)

	recorder = httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "second", recorder.Header().Get("server"))
}

func TestBalancerAllServersFenced(t *testing.T) {
	balancer := New(nil, false)

	balancer.AddServer("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), dynamic.Server{Fenced: true})
	balancer.AddServer("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), dynamic.Server{Fenced: true})

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

// TestSticky_Fenced checks that fenced node receive traffic if their sticky cookie matches.
func TestSticky_Fenced(t *testing.T) {
	balancer := New(&dynamic.Sticky{Cookie: &dynamic.Cookie{Name: "test"}}, false)
	balancer.rand = &mockRand{vals: []int{1, 0, 1, 0}}

	balancer.AddServer("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})

	balancer.AddServer("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{})

	balancer.AddServer("fenced", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fenced")
		rw.WriteHeader(http.StatusOK)
	}), dynamic.Server{Fenced: true})

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}, cookies: make(map[string]*http.Cookie)}

	stickyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	stickyReq.AddCookie(&http.Cookie{Name: "test", Value: "fenced"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for range 2 {
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, stickyReq)
		balancer.ServeHTTP(recorder, req)
	}

	assert.Equal(t, 2, recorder.save["fenced"])
	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 2, recorder.save["second"])
}

func TestStickyWithCompatibility(t *testing.T) {
	testCases := []struct {
		desc    string
		servers []string
		cookies []*http.Cookie

		expectedCookies []*http.Cookie
		expectedServer  string
	}{
		{
			desc:    "No previous cookie",
			servers: []string{"first"},

			expectedServer: "first",
			expectedCookies: []*http.Cookie{
				{Name: "test", Value: loadbalancer.Sha256Hash("first")},
			},
		},
		{
			desc:    "Sha256 previous cookie",
			servers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: loadbalancer.Sha256Hash("first")},
			},
			expectedServer:  "first",
			expectedCookies: []*http.Cookie{},
		},
		{
			desc:    "Raw previous cookie",
			servers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: "first"},
			},
			expectedServer: "first",
			expectedCookies: []*http.Cookie{
				{Name: "test", Value: loadbalancer.Sha256Hash("first")},
			},
		},
		{
			desc:    "Fnv previous cookie",
			servers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: loadbalancer.FnvHash("first")},
			},
			expectedServer: "first",
			expectedCookies: []*http.Cookie{
				{Name: "test", Value: loadbalancer.Sha256Hash("first")},
			},
		},
		{
			desc:    "Double fnv previous cookie",
			servers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: loadbalancer.FnvHash("first")},
			},
			expectedServer: "first",
			expectedCookies: []*http.Cookie{
				{Name: "test", Value: loadbalancer.Sha256Hash("first")},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			balancer := New(&dynamic.Sticky{Cookie: &dynamic.Cookie{Name: "test"}}, false)

			for _, server := range test.servers {
				balancer.AddServer(server, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					_, _ = rw.Write([]byte(server))
				}), dynamic.Server{})
			}

			// Do it twice, to be sure it's not just the luck.
			for range 2 {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				for _, cookie := range test.cookies {
					req.AddCookie(cookie)
				}

				recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}, cookies: make(map[string]*http.Cookie)}
				balancer.ServeHTTP(recorder, req)

				assert.Equal(t, test.expectedServer, recorder.Body.String())

				assert.Len(t, recorder.cookies, len(test.expectedCookies))
				for _, cookie := range test.expectedCookies {
					assert.Equal(t, cookie.Value, recorder.cookies[cookie.Name].Value)
				}
			}
		})
	}
}

type responseRecorder struct {
	*httptest.ResponseRecorder
	save     map[string]int
	sequence []string
	status   []int
	cookies  map[string]*http.Cookie
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.save[r.Header().Get("server")]++
	r.sequence = append(r.sequence, r.Header().Get("server"))
	r.status = append(r.status, statusCode)
	for _, cookie := range r.Result().Cookies() {
		r.cookies[cookie.Name] = cookie
	}
	r.ResponseRecorder.WriteHeader(statusCode)
}

type mockRand struct {
	vals  []int
	calls int
}

func (m *mockRand) Intn(int) int {
	defer func() {
		m.calls++
	}()
	return m.vals[m.calls]
}

func testHandlers(inflights ...int) []*namedHandler {
	var out []*namedHandler
	for i, inflight := range inflights {
		h := &namedHandler{
			name: strconv.Itoa(i),
		}
		h.inflight.Store(int64(inflight))
		out = append(out, h)
	}
	return out
}
