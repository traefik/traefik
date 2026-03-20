package customerrors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestHandler(t *testing.T) {
	testCases := []struct {
		desc                string
		errorPage           *dynamic.ErrorPage
		backendCode         int
		backendHeaders      map[string]string
		backendErrorHandler http.HandlerFunc
		validate            func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			desc:        "no error",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusOK,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusOK))
			},
		},
		{
			desc:        "no error, but not a 200",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusPartialContent,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusPartialContent, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusPartialContent))
			},
		},
		{
			desc:        "a 304, so no Write called",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusNotModified,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, "whatever, should not be called")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusNotModified, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "")
			},
		},
		{
			desc:        "in the range",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusInternalServerError,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusInternalServerError, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My error page.")
			},
		},
		{
			desc:        "not in the range",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusBadGateway,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusBadGateway, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusBadGateway))
			},
		},
		{
			desc:        "query replacement",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/{status}", Status: []string{"503-503"}},
			backendCode: http.StatusServiceUnavailable,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI != "/503" {
					return
				}

				_, _ = fmt.Fprintln(w, "My 503 page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
			},
		},
		{
			desc:        "single code and query replacement",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/{status}", Status: []string{"503"}},
			backendCode: http.StatusServiceUnavailable,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI != "/503" {
					return
				}

				_, _ = fmt.Fprintln(w, "My 503 page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
			},
		},
		{
			desc:        "forward request host header",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/test", Status: []string{"503"}},
			backendCode: http.StatusServiceUnavailable,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, r.Host)
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "localhost")
			},
		},
		{
			desc:        "full query replacement",
			errorPage:   &dynamic.ErrorPage{Service: "error", Query: "/?status={status}&url={url}", Status: []string{"503"}},
			backendCode: http.StatusServiceUnavailable,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI != "/?status=503&url=http%3A%2F%2Flocalhost%2Ftest%3Ffoo%3Dbar%26baz%3Dbuz" {
					t.Log(r.RequestURI)
					return
				}

				_, _ = fmt.Fprintln(w, "My 503 page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
			},
		},
		{
			desc: "nginx headers: backend status code preserved",
			errorPage: &dynamic.ErrorPage{
				Service: "error",
				Query:   "/test",
				Status:  []string{"500-599"},
				NginxHeaders: &http.Header{
					"X-Namespaces":   {"default"},
					"X-Ingress-Name": {"my-ingress"},
					"X-Service-Name": {"my-service"},
					"X-Service-Port": {"80"},
				},
			},
			backendCode: http.StatusInternalServerError,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Error page backend returns 200 (the default when no WriteHeader is called).
				_, _ = fmt.Fprintln(w, "Custom error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				// In nginx mode, the error page backend's status code (200) is preserved,
				// NOT overridden to the original error code (500).
				assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "Custom error page.")
			},
		},
		{
			desc: "nginx headers: X-Code and nginx headers forwarded",
			errorPage: &dynamic.ErrorPage{
				Service: "error",
				Query:   "/test",
				Status:  []string{"500-599"},
				NginxHeaders: &http.Header{
					"X-Namespaces":   {"default"},
					"X-Ingress-Name": {"my-ingress"},
					"X-Service-Name": {"my-service"},
					"X-Service-Port": {"80"},
				},
			},
			backendCode: http.StatusBadGateway,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify that nginx-specific headers are set on the request.
				assert.Equal(t, "502", r.Header.Get("X-Code"))
				assert.Equal(t, "default", r.Header.Get("X-Namespaces"))
				assert.Equal(t, "my-ingress", r.Header.Get("X-Ingress-Name"))
				assert.Equal(t, "my-service", r.Header.Get("X-Service-Name"))
				assert.Equal(t, "80", r.Header.Get("X-Service-Port"))
				assert.Equal(t, "/test?foo=bar&baz=buz", r.Header.Get("X-Original-Uri"))
				// Return a custom status code.
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprintln(w, "Custom 404 page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				// Backend's chosen status code (404) is preserved in nginx mode.
				assert.Equal(t, http.StatusNotFound, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "Custom 404 page.")
			},
		},
		{
			desc: "non-nginx: code modifier enforces original error code",
			errorPage: &dynamic.ErrorPage{
				Service: "error",
				Query:   "/test",
				Status:  []string{"500-599"},
			},
			backendCode: http.StatusInternalServerError,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Error page backend returns 200 (the default).
				_, _ = fmt.Fprintln(w, "Custom error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				// Without nginx headers, newCodeModifier enforces the original error code (500),
				// even though the error page backend returned 200.
				assert.Equal(t, http.StatusInternalServerError, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "Custom error page.")
			},
		},
		{
			desc: "forwardHeaders: WWW-Authenticate forwarded from backend to client",
			errorPage: &dynamic.ErrorPage{
				Service:        "error",
				Query:          "/{status}",
				Status:         []string{"401"},
				ForwardHeaders: []string{"WWW-Authenticate"},
			},
			backendCode: http.StatusUnauthorized,
			backendHeaders: map[string]string{
				"WWW-Authenticate": `Basic realm="Login Required"`,
			},
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, "Error page body.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, recorder.Code, "HTTP status")
				assert.Equal(t, `Basic realm="Login Required"`, recorder.Header().Get("WWW-Authenticate"))
				assert.Contains(t, recorder.Body.String(), "Error page body.")
			},
		},
		{
			desc: "forwardHeaders: headers not in list are not forwarded",
			errorPage: &dynamic.ErrorPage{
				Service:        "error",
				Query:          "/{status}",
				Status:         []string{"500"},
				ForwardHeaders: []string{"WWW-Authenticate"},
			},
			backendCode: http.StatusInternalServerError,
			backendHeaders: map[string]string{
				"X-Custom-Header":  "should-not-appear",
				"WWW-Authenticate": `Bearer realm="example"`,
			},
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, "Error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusInternalServerError, recorder.Code, "HTTP status")
				assert.Equal(t, `Bearer realm="example"`, recorder.Header().Get("WWW-Authenticate"))
				assert.Empty(t, recorder.Header().Get("X-Custom-Header"))
			},
		},
		{
			desc: "forwardHeaders: hop-by-hop headers are filtered out even if listed",
			errorPage: &dynamic.ErrorPage{
				Service:        "error",
				Query:          "/{status}",
				Status:         []string{"401"},
				ForwardHeaders: []string{"WWW-Authenticate", "Connection", "Transfer-Encoding", "Keep-Alive"},
			},
			backendCode: http.StatusUnauthorized,
			backendHeaders: map[string]string{
				"WWW-Authenticate": `Basic realm="test"`,
				"Connection":       "keep-alive",
				"Transfer-Encoding": "chunked",
			},
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, "Error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, recorder.Code, "HTTP status")
				assert.Equal(t, `Basic realm="test"`, recorder.Header().Get("WWW-Authenticate"))
				assert.Empty(t, recorder.Header().Get("Connection"), "hop-by-hop header Connection must not be forwarded")
				assert.Empty(t, recorder.Header().Get("Transfer-Encoding"), "hop-by-hop header Transfer-Encoding must not be forwarded")
			},
		},
		{
			desc: "forwardHeaders: whitespace and duplicates are normalized",
			errorPage: &dynamic.ErrorPage{
				Service:        "error",
				Query:          "/{status}",
				Status:         []string{"401"},
				ForwardHeaders: []string{"  www-authenticate ", "WWW-Authenticate", "x-custom"},
			},
			backendCode: http.StatusUnauthorized,
			backendHeaders: map[string]string{
				"WWW-Authenticate": `Bearer realm="test"`,
				"X-Custom":         "value1",
			},
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, "Error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, recorder.Code, "HTTP status")
				// Despite duplicated WWW-Authenticate in config, only forwarded once.
				assert.Equal(t, `Bearer realm="test"`, recorder.Header().Get("WWW-Authenticate"))
				assert.Equal(t, "value1", recorder.Header().Get("X-Custom"))
			},
		},
		{
			desc: "forwardHeaders not set: backend headers not forwarded (default behavior)",
			errorPage: &dynamic.ErrorPage{
				Service: "error",
				Query:   "/{status}",
				Status:  []string{"401"},
			},
			backendCode: http.StatusUnauthorized,
			backendHeaders: map[string]string{
				"WWW-Authenticate": `Basic realm="Login Required"`,
			},
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, "Error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, recorder.Code, "HTTP status")
				assert.Empty(t, recorder.Header().Get("WWW-Authenticate"))
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			serviceBuilderMock := &mockServiceBuilder{handler: test.backendErrorHandler}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range test.backendHeaders {
					w.Header().Set(k, v)
				}

				w.WriteHeader(test.backendCode)

				if test.backendCode == http.StatusNotModified {
					return
				}
				_, _ = fmt.Fprintln(w, http.StatusText(test.backendCode))
			})
			errorPageHandler, err := New(t.Context(), handler, *test.errorPage, serviceBuilderMock, "test")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost/test?foo=bar&baz=buz", nil)

			// Client like browser and curl will issue a relative HTTP request, which not have a host and scheme in the URL. But the http.NewRequest will set them automatically.
			req.URL.Host = ""
			req.URL.Scheme = ""

			recorder := httptest.NewRecorder()
			errorPageHandler.ServeHTTP(recorder, req)

			test.validate(t, recorder)
		})
	}
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

		h.Add("User-Agent", "foobar")
		_, _ = w.Write([]byte("Hello"))
		w.WriteHeader(http.StatusBadGateway)
	})

	serviceBuilderMock := &mockServiceBuilder{handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "My error page.")
	})}

	config := dynamic.ErrorPage{Service: "error", Query: "/", Status: []string{"200"}}

	errorPageHandler, err := New(t.Context(), next, config, serviceBuilderMock, "test")
	require.NoError(t, err)

	server := httptest.NewServer(errorPageHandler)
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
	assert.Equal(t, "My error page.\n", string(body))
}

type mockServiceBuilder struct {
	handler http.Handler
}

func (m *mockServiceBuilder) BuildHTTP(_ context.Context, _ string) (http.Handler, error) {
	return m.handler, nil
}
