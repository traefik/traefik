package auth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"github.com/vulcand/oxy/v2/forward"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

func TestForwardAuthFail(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(forward.ProxyAuthenticate, "test")
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	middleware, err := NewForward(t.Context(), next, dynamic.ForwardAuth{
		Address: server.URL,
	}, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, "test", res.Header.Get(forward.ProxyAuthenticate))
	assert.Equal(t, "Forbidden\n", string(body))
}

func TestForwardAuthSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-User", "user@example.com")
		w.Header().Set("X-Auth-Secret", "secret")
		w.Header().Add("X-Auth-Group", "group1")
		w.Header().Add("X-Auth-Group", "group2")
		w.Header().Add("Foo-Bar", "auth-value")
		w.Header().Add("Set-Cookie", "authCookie=Auth")
		w.Header().Add("Set-Cookie", "authCookieNotAdded=Auth")
		fmt.Fprintln(w, "Success")
	}))
	t.Cleanup(server.Close)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "user@example.com", r.Header.Get("X-Auth-User"))
		assert.Empty(t, r.Header.Get("X-Auth-Secret"))
		assert.Equal(t, []string{"group1", "group2"}, r.Header["X-Auth-Group"])
		assert.Equal(t, "auth-value", r.Header.Get("Foo-Bar"))
		assert.Empty(t, r.Header.Get("Foo-Baz"))
		w.Header().Add("Set-Cookie", "authCookie=Backend")
		w.Header().Add("Set-Cookie", "backendCookie=Backend")
		w.Header().Add("Other-Header", "BackendHeaderValue")
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.ForwardAuth{
		Address:                  server.URL,
		AuthResponseHeaders:      []string{"X-Auth-User", "X-Auth-Group"},
		AuthResponseHeadersRegex: "^Foo-",
		AddAuthCookiesToResponse: []string{"authCookie"},
	}
	middleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("X-Auth-Group", "admin_group")
	req.Header.Set("Foo-Bar", "client-value")
	req.Header.Set("Foo-Baz", "client-value")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []string{"backendCookie=Backend", "authCookie=Auth"}, res.Header["Set-Cookie"])
	assert.Equal(t, []string{"BackendHeaderValue"}, res.Header["Other-Header"])

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, "traefik\n", string(body))
}

func TestForwardAuthForwardBody(t *testing.T) {
	data := []byte("forwardBodyTest")

	var serverCallCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		serverCallCount++

		forwardedData, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Equal(t, data, forwardedData)
	}))
	t.Cleanup(server.Close)

	var nextCallCount int
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		nextCallCount++
	})

	maxBodySize := int64(len(data))
	auth := dynamic.ForwardAuth{Address: server.URL, ForwardBody: true, MaxBodySize: &maxBodySize}

	middleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, bytes.NewReader(data))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, 1, serverCallCount)
	assert.Equal(t, 1, nextCallCount)
}

func TestForwardAuthForwardBodyEmptyBody(t *testing.T) {
	var serverCallCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		serverCallCount++

		forwardedData, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Empty(t, forwardedData)
	}))
	t.Cleanup(server.Close)

	var nextCallCount int
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		nextCallCount++
	})

	auth := dynamic.ForwardAuth{Address: server.URL, ForwardBody: true}

	middleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, http.NoBody)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, 1, serverCallCount)
	assert.Equal(t, 1, nextCallCount)
}

func TestForwardAuthForwardBodySizeLimit(t *testing.T) {
	data := []byte("forwardBodyTest")

	var serverCallCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		serverCallCount++

		forwardedData, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Equal(t, data, forwardedData)
	}))
	t.Cleanup(server.Close)

	var nextCallCount int
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		nextCallCount++
	})

	maxBodySize := int64(len(data)) - 1
	auth := dynamic.ForwardAuth{Address: server.URL, ForwardBody: true, MaxBodySize: &maxBodySize}

	middleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, bytes.NewReader(data))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.Equal(t, 0, serverCallCount)
	assert.Equal(t, 0, nextCallCount)
}

func TestForwardAuthNotForwardBody(t *testing.T) {
	data := []byte("forwardBodyTest")

	var serverCallCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		serverCallCount++

		forwardedData, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Empty(t, forwardedData)
	}))
	t.Cleanup(server.Close)

	var nextCallCount int
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		nextCallCount++
	})

	auth := dynamic.ForwardAuth{Address: server.URL}

	middleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, bytes.NewReader(data))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, 1, serverCallCount)
	assert.Equal(t, 1, nextCallCount)
}

func TestForwardAuthRedirect(t *testing.T) {
	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://example.com/redirect-test", http.StatusFound)
	}))
	t.Cleanup(authTs.Close)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.ForwardAuth{Address: authTs.URL}

	authMiddleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	t.Cleanup(ts.Close)

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)

	res, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusFound, res.StatusCode)

	location, err := res.Location()
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/redirect-test", location.String())

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	assert.NotEmpty(t, string(body))
}

func TestForwardAuthRemoveHopByHopHeaders(t *testing.T) {
	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		for _, header := range hopHeaders {
			if header == forward.TransferEncoding {
				headers.Set(header, "chunked")
			} else {
				headers.Add(header, "test")
			}
		}

		http.Redirect(w, r, "http://example.com/redirect-test", http.StatusFound)
	}))
	t.Cleanup(authTs.Close)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.ForwardAuth{Address: authTs.URL}

	authMiddleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	t.Cleanup(ts.Close)

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, res.StatusCode, "they should be equal")

	for _, header := range forward.HopHeaders {
		assert.Empty(t, res.Header.Get(header), "hop-by-hop header '%s' mustn't be set", header)
	}

	location, err := res.Location()
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/redirect-test", location.String(), "they should be equal")

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.NotEmpty(t, string(body), "there should be something in the body")
}

func TestForwardAuthFailResponseHeaders(t *testing.T) {
	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := &http.Cookie{Name: "example", Value: "testing", Path: "/"}
		http.SetCookie(w, cookie)
		w.Header().Add("X-Foo", "bar")
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	t.Cleanup(authTs.Close)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.ForwardAuth{
		Address: authTs.URL,
	}
	authMiddleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	require.Len(t, res.Cookies(), 1)
	for _, cookie := range res.Cookies() {
		assert.Equal(t, "testing", cookie.Value)
	}

	expectedHeaders := http.Header{
		"Content-Length":         []string{"10"},
		"Content-Type":           []string{"text/plain; charset=utf-8"},
		"X-Foo":                  []string{"bar"},
		"Set-Cookie":             []string{"example=testing; Path=/"},
		"X-Content-Type-Options": []string{"nosniff"},
	}

	assert.Len(t, res.Header, 6)
	for key, value := range expectedHeaders {
		assert.Equal(t, value, res.Header[key])
	}

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, "Forbidden\n", string(body))
}

func TestForwardAuthClientClosedRequest(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCancelled := make(chan struct{})
	responseComplete := make(chan struct{})

	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-requestCancelled
	}))
	t.Cleanup(authTs.Close)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// next should not be called.
		t.Fail()
	})

	auth := dynamic.ForwardAuth{
		Address: authTs.URL,
	}
	authMiddleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequestWithContext(ctx, "GET", "http://foo", http.NoBody)

	recorder := httptest.NewRecorder()
	go func() {
		authMiddleware.ServeHTTP(recorder, req)
		close(responseComplete)
	}()

	<-requestStarted

	cancel()
	close(requestCancelled)

	<-responseComplete

	assert.Equal(t, httputil.StatusClientClosedRequest, recorder.Result().StatusCode)
}

func TestForwardAuthForwardError(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// next should not be called.
		t.Fail()
	})

	auth := dynamic.ForwardAuth{
		Address: "http://non-existing-server",
	}
	authMiddleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Microsecond)
	defer cancel()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://foo", nil)

	recorder := httptest.NewRecorder()
	responseComplete := make(chan struct{})
	go func() {
		authMiddleware.ServeHTTP(recorder, req)
		close(responseComplete)
	}()

	<-responseComplete

	assert.Equal(t, http.StatusInternalServerError, recorder.Result().StatusCode)
}

func Test_writeHeader(t *testing.T) {
	testCases := []struct {
		name                      string
		headers                   map[string]string
		authRequestHeaders        []string
		trustForwardHeader        bool
		emptyHost                 bool
		expectedHeaders           map[string]string
		checkForUnexpectedHeaders bool
	}{
		{
			name: "trust Forward Header",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
		},
		{
			name: "not trust Forward Header",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "foo.bar",
			},
		},
		{
			name: "trust Forward Header with empty Host",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: true,
			emptyHost:          true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
		},
		{
			name: "not trust Forward Header with empty Host",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: false,
			emptyHost:          true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "",
			},
		},
		{
			name: "trust Forward Header with forwarded URI",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
				"X-Forwarded-Uri":  "/forward?q=1",
			},
			trustForwardHeader: true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
				"X-Forwarded-Uri":  "/forward?q=1",
			},
		},
		{
			name: "not trust Forward Header with forward requested URI",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
				"X-Forwarded-Uri":  "/forward?q=1",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "foo.bar",
				"X-Forwarded-Uri":  "/path?q=1",
			},
		},
		{
			name: "trust Forward Header with forwarded request Method",
			headers: map[string]string{
				"X-Forwarded-Method": "OPTIONS",
			},
			trustForwardHeader: true,
			expectedHeaders: map[string]string{
				"X-Forwarded-Method": "OPTIONS",
			},
		},
		{
			name: "not trust Forward Header with forward request Method",
			headers: map[string]string{
				"X-Forwarded-Method": "OPTIONS",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"X-Forwarded-Method": "GET",
			},
		},
		{
			name: "remove hop-by-hop headers",
			headers: map[string]string{
				forward.Connection:         "Connection",
				forward.KeepAlive:          "KeepAlive",
				forward.ProxyAuthenticate:  "ProxyAuthenticate",
				forward.ProxyAuthorization: "ProxyAuthorization",
				forward.Te:                 "Te",
				forward.Trailers:           "Trailers",
				forward.TransferEncoding:   "TransferEncoding",
				forward.Upgrade:            "Upgrade",
				"X-CustomHeader":           "CustomHeader",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"X-CustomHeader":           "CustomHeader",
				"X-Forwarded-Proto":        "http",
				"X-Forwarded-Host":         "foo.bar",
				"X-Forwarded-Uri":          "/path?q=1",
				"X-Forwarded-Method":       "GET",
				forward.ProxyAuthenticate:  "ProxyAuthenticate",
				forward.ProxyAuthorization: "ProxyAuthorization",
			},
			checkForUnexpectedHeaders: true,
		},
		{
			name: "filter forward request headers",
			headers: map[string]string{
				"X-CustomHeader": "CustomHeader",
				"Content-Type":   "multipart/form-data; boundary=---123456",
			},
			authRequestHeaders: []string{
				"X-CustomHeader",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"x-customHeader":     "CustomHeader",
				"X-Forwarded-Proto":  "http",
				"X-Forwarded-Host":   "foo.bar",
				"X-Forwarded-Uri":    "/path?q=1",
				"X-Forwarded-Method": "GET",
			},
			checkForUnexpectedHeaders: true,
		},
		{
			name: "filter forward request headers doesn't add new headers",
			headers: map[string]string{
				"X-CustomHeader": "CustomHeader",
				"Content-Type":   "multipart/form-data; boundary=---123456",
			},
			authRequestHeaders: []string{
				"X-CustomHeader",
				"X-Non-Exists-Header",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"X-CustomHeader":     "CustomHeader",
				"X-Forwarded-Proto":  "http",
				"X-Forwarded-Host":   "foo.bar",
				"X-Forwarded-Uri":    "/path?q=1",
				"X-Forwarded-Method": "GET",
			},
			checkForUnexpectedHeaders: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/path?q=1", nil)
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			if test.emptyHost {
				req.Host = ""
			}

			forwardReq := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/path?q=1", nil)

			writeHeader(req, forwardReq, test.trustForwardHeader, test.authRequestHeaders)

			actualHeaders := forwardReq.Header

			expectedHeaders := test.expectedHeaders
			for key, value := range expectedHeaders {
				assert.Equal(t, value, actualHeaders.Get(key))
				actualHeaders.Del(key)
			}
			if test.checkForUnexpectedHeaders {
				for key := range actualHeaders {
					assert.Fail(t, "Unexpected header found", key)
				}
			}
		})
	}
}

func TestForwardAuthTracing(t *testing.T) {
	type expected struct {
		name       string
		attributes []attribute.KeyValue
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Traceparent") == "" {
			t.Errorf("expected Traceparent header to be present in request")
		}

		w.Header().Set("X-Bar", "foo")
		w.Header().Add("X-Bar", "bar")
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	parse, err := url.Parse(server.URL)
	require.NoError(t, err)

	_, serverPort, err := net.SplitHostPort(parse.Host)
	require.NoError(t, err)

	serverPortInt, err := strconv.Atoi(serverPort)
	require.NoError(t, err)

	testCases := []struct {
		desc     string
		expected []expected
	}{
		{
			desc: "basic test",
			expected: []expected{
				{
					name: "initial",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "unspecified"),
					},
				},
				{
					name: "AuthRequest",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "client"),
						attribute.String("http.request.method", "GET"),
						attribute.String("network.protocol.version", "1.1"),
						attribute.String("url.full", server.URL),
						attribute.String("url.scheme", "http"),
						attribute.String("user_agent.original", ""),
						attribute.String("network.peer.address", "127.0.0.1"),
						attribute.Int64("network.peer.port", int64(serverPortInt)),
						attribute.String("server.address", "127.0.0.1"),
						attribute.Int64("server.port", int64(serverPortInt)),
						attribute.StringSlice("http.request.header.x-foo", []string{"foo", "bar"}),
						attribute.Int64("http.response.status_code", int64(404)),
						attribute.StringSlice("http.response.header.x-bar", []string{"foo", "bar"}),
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

			auth := dynamic.ForwardAuth{
				Address:            server.URL,
				AuthRequestHeaders: []string{"X-Foo"},
			}
			next, err := NewForward(t.Context(), next, auth, "authTest")
			require.NoError(t, err)

			next = observability.WithObservabilityHandler(next, observability.Observability{
				TracingEnabled: true,
			})

			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/search?q=Opentelemetry", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "forward-test")
			req.Header.Set("X-Forwarded-Proto", "http")
			req.Header.Set("X-Foo", "foo")
			req.Header.Add("X-Foo", "bar")

			otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

			mockTracer := &mockTracer{}
			tracer := tracing.NewTracer(mockTracer, []string{"X-Foo"}, []string{"X-Bar"}, []string{"q"})
			initialCtx, initialSpan := tracer.Start(req.Context(), "initial")
			defer initialSpan.End()
			req = req.WithContext(initialCtx)

			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, req)

			for i, span := range mockTracer.spans {
				assert.Equal(t, test.expected[i].name, span.name)
				assert.Equal(t, test.expected[i].attributes, span.attributes)
			}
		})
	}
}

func TestForwardAuthPreserveLocationHeader(t *testing.T) {
	relativeURL := "/index.html"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", relativeURL)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	auth := dynamic.ForwardAuth{
		Address:                server.URL,
		PreserveLocationHeader: true,
	}
	middleware, err := NewForward(t.Context(), next, auth, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	t.Cleanup(ts.Close)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.Equal(t, relativeURL, res.Header.Get("Location"))
}

func TestForwardAuthPreserveRequestMethod(t *testing.T) {
	testCases := []struct {
		name                          string
		preserveRequestMethod         bool
		originalReqMethod             string
		expectedReqMethodInAuthServer string
	}{
		{
			name:                          "preserve request method",
			preserveRequestMethod:         true,
			originalReqMethod:             http.MethodPost,
			expectedReqMethodInAuthServer: http.MethodPost,
		},
		{
			name:                          "do not preserve request method",
			preserveRequestMethod:         false,
			originalReqMethod:             http.MethodPost,
			expectedReqMethodInAuthServer: http.MethodGet,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			reqReachesAuthServer := false
			authServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				reqReachesAuthServer = true
				assert.Equal(t, test.expectedReqMethodInAuthServer, req.Method)
			}))
			t.Cleanup(authServer.Close)

			reqReachesNextServer := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reqReachesNextServer = true
				assert.Equal(t, test.originalReqMethod, r.Method)
			})

			auth := dynamic.ForwardAuth{
				Address:               authServer.URL,
				PreserveRequestMethod: test.preserveRequestMethod,
			}

			middleware, err := NewForward(t.Context(), next, auth, "authTest")
			require.NoError(t, err)

			ts := httptest.NewServer(middleware)
			t.Cleanup(ts.Close)

			req := testhelpers.MustNewRequest(test.originalReqMethod, ts.URL, nil)
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.True(t, reqReachesAuthServer)
			assert.True(t, reqReachesNextServer)
		})
	}
}

type mockTracer struct {
	embedded.Tracer

	spans []*mockSpan
}

var _ trace.Tracer = &mockTracer{}

func (t *mockTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	config := trace.NewSpanStartConfig(opts...)
	span := &mockSpan{}
	span.SetName(name)
	span.SetAttributes(attribute.String("span.kind", config.SpanKind().String()))
	span.SetAttributes(config.Attributes()...)
	t.spans = append(t.spans, span)
	return trace.ContextWithSpan(ctx, span), span
}

// mockSpan is an implementation of Span that preforms no operations.
type mockSpan struct {
	embedded.Span

	name       string
	attributes []attribute.KeyValue
}

var _ trace.Span = &mockSpan{}

func (*mockSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}})
}
func (*mockSpan) IsRecording() bool                  { return false }
func (s *mockSpan) SetStatus(_ codes.Code, _ string) {}
func (s *mockSpan) SetAttributes(kv ...attribute.KeyValue) {
	s.attributes = append(s.attributes, kv...)
}
func (s *mockSpan) End(...trace.SpanEndOption)                  {}
func (s *mockSpan) RecordError(_ error, _ ...trace.EventOption) {}
func (s *mockSpan) AddEvent(_ string, _ ...trace.EventOption)   {}
func (s *mockSpan) AddLink(_ trace.Link)                        {}

func (s *mockSpan) SetName(name string) { s.name = name }

func (s *mockSpan) TracerProvider() trace.TracerProvider {
	return nil
}
