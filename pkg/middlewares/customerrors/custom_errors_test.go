package customerrors

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		test := test
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
			errorPageHandler, err := New(context.Background(), handler, *test.errorPage, serviceBuilderMock, "test")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost/test?foo=bar&baz=buz", nil)

			recorder := httptest.NewRecorder()
			errorPageHandler.ServeHTTP(recorder, req)

			test.validate(t, recorder)
		})
	}
}

type mockServiceBuilder struct {
	handler http.Handler
}

func (m *mockServiceBuilder) BuildHTTP(_ context.Context, _ string) (http.Handler, error) {
	return m.handler, nil
}
