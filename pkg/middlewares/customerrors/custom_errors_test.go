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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			serviceBuilderMock := &mockServiceBuilder{handler: test.backendErrorHandler}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.backendCode)

				if test.backendCode == http.StatusNotModified {
					return
				}
				_, _ = fmt.Fprintln(w, http.StatusText(test.backendCode))
			})
			errorPageHandler, err := New(t.Context(), handler, *test.errorPage, serviceBuilderMock, "test")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost/test?foo=bar&baz=buz", nil)

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
