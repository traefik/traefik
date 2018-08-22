package errorpages

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
)

func TestHandler(t *testing.T) {
	testCases := []struct {
		desc                string
		errorPage           *types.ErrorPage
		backendCode         int
		backendErrorHandler http.HandlerFunc
		validate            func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			desc:        "no error",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusOK,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusOK))
			},
		},
		{
			desc:        "in the range",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusInternalServerError,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My error page.")
				assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
			},
		},
		{
			desc:        "not in the range",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusBadGateway,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadGateway, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusBadGateway))
				assert.NotContains(t, recorder.Body.String(), "Test Server", "Should return the oops page since we have not configured the 502 code")
			},
		},
		{
			desc:        "query replacement",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503-503"}},
			backendCode: http.StatusServiceUnavailable,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == "/503" {
					fmt.Fprintln(w, "My 503 page.")
				} else {
					fmt.Fprintln(w, "Failed")
				}
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
				assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
			},
		},
		{
			desc:        "Single code",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503"}},
			backendCode: http.StatusServiceUnavailable,
			backendErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == "/503" {
					fmt.Fprintln(w, "My 503 page.")
				} else {
					fmt.Fprintln(w, "Failed")
				}
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
				assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			errorPageHandler, err := NewHandler(test.errorPage, "test")
			require.NoError(t, err)

			errorPageHandler.backendHandler = test.backendErrorHandler

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.backendCode)
				fmt.Fprintln(w, http.StatusText(test.backendCode))
			})

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost/test", nil)

			n := negroni.New()
			n.Use(errorPageHandler)
			n.UseHandler(handler)

			recorder := httptest.NewRecorder()
			n.ServeHTTP(recorder, req)

			test.validate(t, recorder)
		})
	}
}

func TestHandlerOldWay(t *testing.T) {
	testCases := []struct {
		desc               string
		errorPage          *types.ErrorPage
		backendCode        int
		errorPageForwarder http.HandlerFunc
		validate           func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			desc:        "no error",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusOK,
			errorPageForwarder: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "OK")
			},
		},
		{
			desc:        "in the range",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusInternalServerError,
			errorPageForwarder: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assert.Contains(t, recorder.Body.String(), "My error page.")
				assert.NotContains(t, recorder.Body.String(), http.StatusText(http.StatusInternalServerError), "Should not return the oops page")
			},
		},
		{
			desc:        "not in the range",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusBadGateway,
			errorPageForwarder: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "My error page.")
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadGateway, recorder.Code)
				assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusBadGateway))
				assert.NotContains(t, recorder.Body.String(), "My error page.", "Should return the oops page since we have not configured the 502 code")
			},
		},
		{
			desc:        "query replacement",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503-503"}},
			backendCode: http.StatusServiceUnavailable,
			errorPageForwarder: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RequestURI() == "/"+strconv.Itoa(503) {
					fmt.Fprintln(w, "My 503 page.")
				} else {
					fmt.Fprintln(w, "Failed")
				}
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
				assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
			},
		},
		{
			desc:        "Single code",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503"}},
			backendCode: http.StatusServiceUnavailable,
			errorPageForwarder: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RequestURI() == "/"+strconv.Itoa(503) {
					fmt.Fprintln(w, "My 503 page.")
				} else {
					fmt.Fprintln(w, "Failed")
				}
			}),
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
				assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
			},
		},
	}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost/test", nil)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			errorPageHandler, err := NewHandler(test.errorPage, "test")
			require.NoError(t, err)
			errorPageHandler.FallbackURL = "http://localhost"

			errorPageHandler.PostLoad(test.errorPageForwarder)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.backendCode)
				fmt.Fprintln(w, http.StatusText(test.backendCode))
			})

			n := negroni.New()
			n.Use(errorPageHandler)
			n.UseHandler(handler)

			recorder := httptest.NewRecorder()
			n.ServeHTTP(recorder, req)

			test.validate(t, recorder)
		})
	}
}

func TestHandlerOldWayIntegration(t *testing.T) {
	errorPagesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/503" {
			fmt.Fprintln(w, "My 503 page.")
		} else {
			fmt.Fprintln(w, "Test Server")
		}
	}))
	defer errorPagesServer.Close()

	testCases := []struct {
		desc        string
		errorPage   *types.ErrorPage
		backendCode int
		validate    func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			desc:        "no error",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusOK,
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "OK")
			},
		},
		{
			desc:        "in the range",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusInternalServerError,
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assert.Contains(t, recorder.Body.String(), "Test Server")
				assert.NotContains(t, recorder.Body.String(), http.StatusText(http.StatusInternalServerError), "Should not return the oops page")
			},
		},
		{
			desc:        "not in the range",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}},
			backendCode: http.StatusBadGateway,
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadGateway, recorder.Code)
				assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusBadGateway))
				assert.NotContains(t, recorder.Body.String(), "Test Server", "Should return the oops page since we have not configured the 502 code")
			},
		},
		{
			desc:        "query replacement",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503-503"}},
			backendCode: http.StatusServiceUnavailable,
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
				assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
			},
		},
		{
			desc:        "Single code",
			errorPage:   &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503"}},
			backendCode: http.StatusServiceUnavailable,
			validate: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status")
				assert.Contains(t, recorder.Body.String(), "My 503 page.")
				assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
			},
		},
	}

	req := testhelpers.MustNewRequest(http.MethodGet, errorPagesServer.URL+"/test", nil)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {

			errorPageHandler, err := NewHandler(test.errorPage, "test")
			require.NoError(t, err)
			errorPageHandler.FallbackURL = errorPagesServer.URL

			err = errorPageHandler.PostLoad(nil)
			require.NoError(t, err)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.backendCode)
				fmt.Fprintln(w, http.StatusText(test.backendCode))
			})

			n := negroni.New()
			n.Use(errorPageHandler)
			n.UseHandler(handler)

			recorder := httptest.NewRecorder()
			n.ServeHTTP(recorder, req)

			test.validate(t, recorder)
		})
	}
}

func TestNewResponseRecorder(t *testing.T) {
	testCases := []struct {
		desc     string
		rw       http.ResponseWriter
		expected http.ResponseWriter
	}{
		{
			desc:     "Without Close Notify",
			rw:       httptest.NewRecorder(),
			expected: &responseRecorderWithoutCloseNotify{},
		},
		{
			desc:     "With Close Notify",
			rw:       &mockRWCloseNotify{},
			expected: &responseRecorderWithCloseNotify{},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rec := newResponseRecorder(test.rw)

			assert.IsType(t, rec, test.expected)
		})
	}
}

type mockRWCloseNotify struct{}

func (m *mockRWCloseNotify) CloseNotify() <-chan bool {
	panic("implement me")
}

func (m *mockRWCloseNotify) Header() http.Header {
	panic("implement me")
}

func (m *mockRWCloseNotify) Write([]byte) (int, error) {
	panic("implement me")
}

func (m *mockRWCloseNotify) WriteHeader(int) {
	panic("implement me")
}
