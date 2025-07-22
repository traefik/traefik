package p2c

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestP2C(t *testing.T) {
	testCases := []struct {
		desc            string
		handlers        []*namedHandler
		rand            *mockRand
		expectedHandler string
	}{
		{
			desc:            "one healthy handler",
			handlers:        testHandlers(0),
			rand:            nil,
			expectedHandler: "0",
		},
		{
			desc:            "two handlers zero in flight",
			handlers:        testHandlers(0, 0),
			rand:            &mockRand{vals: []int{1, 0}},
			expectedHandler: "1",
		},
		{
			desc:            "chooses lower of two",
			handlers:        testHandlers(0, 1),
			rand:            &mockRand{vals: []int{1, 0}},
			expectedHandler: "0",
		},
		{
			desc:            "chooses lower of three",
			handlers:        testHandlers(10, 90, 40),
			rand:            &mockRand{vals: []int{1, 1}},
			expectedHandler: "2",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			balancer := New(nil, false)
			balancer.rand = test.rand

			for _, h := range test.handlers {
				balancer.handlers = append(balancer.handlers, h)
				balancer.status[h.name] = struct{}{}
			}

			got, err := balancer.nextServer()
			require.NoError(t, err)

			assert.Equal(t, test.expectedHandler, got.name)
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

func TestSticky_Fallback(t *testing.T) {
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
	balancer.SetStatus(t.Context(), "second", false)
	assert.Equal(t, 0, calls)

	recorder = httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "first", recorder.Header().Get("server"))

	// first gets downed, balancer is down.
	balancer.SetStatus(t.Context(), "first", false)
	assert.Equal(t, 1, calls)

	recorder = httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)

	// two gets up, balancer up.
	balancer.SetStatus(t.Context(), "second", true)
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
