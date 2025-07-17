package headers

// Middleware tests based on https://github.com/unrolled/secure

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestNew_withoutOptions(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	mid, err := New(t.Context(), next, dynamic.Headers{}, "testing")
	require.Errorf(t, err, "headers configuration not valid")

	assert.Nil(t, mid)
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
