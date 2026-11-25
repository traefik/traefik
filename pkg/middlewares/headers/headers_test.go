package headers

// Middleware tests based on https://github.com/unrolled/secure

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/label"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
)

func TestNew_withoutOptions(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	mid, err := New(t.Context(), next, dynamic.Headers{}, "testing")
	require.Errorf(t, err, "headers configuration not valid")

	assert.Nil(t, mid)
}

func TestNew_withSTSSecondsZero(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	mid, err := New(t.Context(), next, dynamic.Headers{STSSeconds: new(int64(0))}, "testing")
	require.NoError(t, err)
	assert.NotNil(t, mid)
}

func TestNew_allowedHosts(t *testing.T) {
	testCases := []struct {
		desc     string
		fromHost string
		expected int
	}{
		{
			desc:     "Should accept the request when given a host that is in the list",
			fromHost: "foo.com",
			expected: http.StatusOK,
		},
		{
			desc:     "Should refuse the request when no host is given",
			fromHost: "",
			expected: http.StatusInternalServerError,
		},
		{
			desc:     "Should refuse the request when no matching host is given",
			fromHost: "boo.com",
			expected: http.StatusInternalServerError,
		},
	}

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	cfg := dynamic.Headers{
		AllowedHosts: []string{"foo.com", "bar.com"},
	}

	mid, err := New(t.Context(), emptyHandler, cfg, "foo")
	require.NoError(t, err)

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Host = test.fromHost

			rw := httptest.NewRecorder()

			mid.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Code)
		})
	}
}

func TestNew_customHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	cfg := dynamic.Headers{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
		CustomResponseHeaders: map[string]string{
			"X-Custom-Response-Header": "test_response",
		},
	}

	mid, err := New(t.Context(), next, cfg, "testing")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	rw := httptest.NewRecorder()

	mid.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"))
	assert.Equal(t, "test_response", rw.Header().Get("X-Custom-Response-Header"))
}

func TestNew_headerOperations(t *testing.T) {
	var middleware traefikv1alpha1.Middleware
	err := json.Unmarshal([]byte(`{
		"spec": {
			"headers": {
				"customRequestHeaders": {"X-Legacy": "request"},
				"customResponseHeaders": {"X-Legacy": "response"},
				"requestHeader": {
					"set": {"Host": "backend.example", "X-Set": "set"},
					"append": {"x-values": "third", "X-Set": "appended", "X-Delete": "appended"},
					"delete": ["x-delete"]
				},
				"responseHeader": {
					"set": {"X-Set": "set"},
					"append": {"set-cookie": "theme=dark", "X-New": "new", "X-Set": "appended", "X-Delete": "appended"},
					"delete": ["x-delete"]
				}
			}
		}
	}`), &middleware)
	require.NoError(t, err)

	cfg := (&dynamic.Middleware{Headers: middleware.Spec.Headers}).DeepCopy()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, []string{"first, second", "original", "third"}, r.Header.Values("X-Values"))
		assert.Equal(t, []string{"set", "appended"}, r.Header.Values("X-Set"))
		assert.NotContains(t, r.Header, "X-Delete")
		assert.Equal(t, "request", r.Header.Get("X-Legacy"))
		assert.Equal(t, "backend.example", r.Host)

		w.Header().Add("Set-Cookie", "session=abc")
		w.Header().Add("Set-Cookie", "lang=en")
		w.Header().Set("X-Set", "original")
		w.Header().Set("X-Delete", "original")
		w.WriteHeader(http.StatusNoContent)
	})

	mid, err := New(t.Context(), next, *cfg.Headers, "testing")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://original.example", nil)
	req.Header.Add("X-Values", "first, second")
	req.Header.Add("X-Values", "original")
	req.Header.Set("X-Set", "original")
	req.Header.Set("X-Delete", "original")
	rw := httptest.NewRecorder()
	mid.ServeHTTP(rw, req)

	res := rw.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.Equal(t, []string{"session=abc", "lang=en", "theme=dark"}, res.Header.Values("Set-Cookie"))
	assert.Equal(t, []string{"set", "appended"}, res.Header.Values("X-Set"))
	assert.Equal(t, "new", res.Header.Get("X-New"))
	assert.NotContains(t, res.Header, "X-Delete")
	assert.Equal(t, "response", res.Header.Get("X-Legacy"))
}

func TestNew_appendHeadersFromLabels(t *testing.T) {
	cfg, err := label.DecodeConfiguration(map[string]string{
		"traefik.http.middlewares.test.headers.requestheader.append.X-Feature":      "enabled",
		"traefik.http.middlewares.test.headers.responseheader.append.Cache-Control": "private",
	})
	require.NoError(t, err)

	copied := cfg.DeepCopy()
	cfg.HTTP.Middlewares["test"].Headers.RequestHeader.Append["X-Feature"] = "changed"
	cfg.HTTP.Middlewares["test"].Headers.ResponseHeader.Append["Cache-Control"] = "changed"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, []string{"enabled"}, r.Header.Values("X-Feature"))
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusNoContent)
	})

	mid, err := New(t.Context(), next, *copied.HTTP.Middlewares["test"].Headers, "test")
	require.NoError(t, err)

	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rw := httptest.NewRecorder()
		mid.ServeHTTP(rw, req)

		res := rw.Result()
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.Equal(t, []string{"max-age=60", "private"}, res.Header.Values("Cache-Control"))
		require.NoError(t, res.Body.Close())
	}
}

func TestNew_appendVaryHeaderWithCORS(t *testing.T) {
	// Spare capacity lets Header.Add reuse the backend's slice.
	backendVary := make([]string, 2, 3)
	copy(backendVary, []string{"Accept-Encoding", "Accept-Language"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header()["Vary"] = backendVary
		w.WriteHeader(http.StatusNoContent)
	})

	cfg := dynamic.Headers{
		AddVaryHeader:  true,
		ResponseHeader: &dynamic.HeaderOperations{Append: map[string]string{"Vary": "Accept"}},
	}
	mid, err := New(t.Context(), next, cfg, "test")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	mid.ServeHTTP(rw, req)

	res := rw.Result()
	defer res.Body.Close()
	assert.Equal(t, []string{"Accept-Encoding,Origin", "Accept-Language", "Accept"}, res.Header.Values("Vary"))
	assert.Equal(t, []string{"Accept-Encoding", "Accept-Language"}, backendVary)
}

func Test_headers_getTracingInformation(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mid := &headers{
		handler: next,
		name:    "testing",
	}

	name, typeName := mid.GetTracingInformation()

	assert.Equal(t, "testing", name)
	assert.Equal(t, "Headers", typeName)
}

// This test is an adapted version of net/http/httputil.Test1xxResponses test.
func Test1xxResponses(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Link", "</style.css>; rel=preload; as=style")
		h.Add("Link", "</script.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusEarlyHints)

		h.Add("Link", "</foo.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusProcessing)

		_, _ = w.Write([]byte("Hello"))
	})

	cfg := dynamic.Headers{
		CustomResponseHeaders: map[string]string{
			"X-Custom-Response-Header": "test_response",
		},
	}

	mid, err := New(t.Context(), next, cfg, "testing")
	require.NoError(t, err)

	server := httptest.NewServer(mid)
	t.Cleanup(server.Close)
	frontendClient := server.Client()

	checkLinkHeaders := func(t *testing.T, expected, got []string) {
		t.Helper()

		if len(expected) != len(got) {
			t.Errorf("Expected %d link headers; got %d", len(expected), len(got))
		}

		for i := range expected {
			if i >= len(got) {
				t.Errorf("Expected %q link header; got nothing", expected[i])

				continue
			}

			if expected[i] != got[i] {
				t.Errorf("Expected %q link header; got %q", expected[i], got[i])
			}
		}
	}

	var respCounter uint8
	trace := &httptrace.ClientTrace{
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			switch code {
			case http.StatusEarlyHints:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script"}, header["Link"])
			case http.StatusProcessing:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, header["Link"])
			default:
				t.Error("Unexpected 1xx response")
			}

			respCounter++

			return nil
		},
	}
	req, _ := http.NewRequestWithContext(httptrace.WithClientTrace(t.Context(), trace), http.MethodGet, server.URL, nil)

	res, err := frontendClient.Do(req)
	assert.NoError(t, err)

	defer res.Body.Close()

	if respCounter != 2 {
		t.Errorf("Expected 2 1xx responses; got %d", respCounter)
	}
	checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, res.Header["Link"])

	body, _ := io.ReadAll(res.Body)
	if string(body) != "Hello" {
		t.Errorf("Read body %q; want Hello", body)
	}

	assert.Equal(t, "test_response", res.Header.Get("X-Custom-Response-Header"))
}
