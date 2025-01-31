package integration

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

// Headers tests suite.
type HeadersSuite struct{ BaseSuite }

func TestHeadersSuite(t *testing.T) {
	suite.Run(t, new(HeadersSuite))
}

func (s *HeadersSuite) TearDownTest() {
	s.displayTraefikLogFile(traefikTestLogFile)
	_ = os.Remove(traefikTestAccessLogFile)
}

func (s *HeadersSuite) TestSimpleConfiguration() {
	s.traefikCmd(withConfigFile("fixtures/headers/basic.toml"))

	// Expected a 404 as we did not configure anything
	err := try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *HeadersSuite) TestReverseProxyHeaderRemoved() {
	file := s.adaptFile("fixtures/headers/remove_reverseproxy_headers.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found := r.Header["X-Forwarded-Host"]
		assert.True(s.T(), found)
		_, found = r.Header["Foo"]
		assert.False(s.T(), found)
		_, found = r.Header["X-Forwarded-For"]
		assert.False(s.T(), found)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	require.NoError(s.T(), err)

	ts := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	ts.Start()
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "test.localhost"
	req.Header = http.Header{
		"Foo": {"bar"},
	}

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *HeadersSuite) TestConnectionHopByHop() {
	file := s.adaptFile("fixtures/headers/connection_hop_by_hop_headers.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found := r.Header["X-Forwarded-For"]
		assert.True(s.T(), found)
		xHost, found := r.Header["X-Forwarded-Host"]
		assert.True(s.T(), found)
		assert.Equal(s.T(), "localhost", xHost[0])

		_, found = r.Header["Foo"]
		assert.False(s.T(), found)
		_, found = r.Header["Bar"]
		assert.False(s.T(), found)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	require.NoError(s.T(), err)

	ts := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	ts.Start()
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "test.localhost"
	req.Header = http.Header{
		"Connection":       {"Foo,Bar,X-Forwarded-For,X-Forwarded-Host"},
		"Foo":              {"bar"},
		"Bar":              {"foo"},
		"X-Forwarded-Host": {"localhost"},
	}

	err = try.Request(req, time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	accessLog, err := os.ReadFile(traefikTestAccessLogFile)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(accessLog), "\"request_Foo\":\"bar\"")
	assert.NotContains(s.T(), string(accessLog), "\"request_Bar\":\"\"")
}

func (s *HeadersSuite) TestCorsResponses() {
	file := s.adaptFile("fixtures/headers/cors.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	backend := startTestServer("9000", http.StatusOK, "")
	defer backend.Close()

	err := try.GetRequest(backend.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	testCase := []struct {
		desc           string
		requestHeaders http.Header
		expected       http.Header
		reqHost        string
		method         string
	}{
		{
			desc: "simple access control allow origin",
			requestHeaders: http.Header{
				"Origin": {"https://foo.bar.org"},
			},
			expected: http.Header{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
				"Vary":                        {"Origin"},
			},
			reqHost: "test.localhost",
			method:  http.MethodGet,
		},
		{
			desc: "simple preflight request",
			requestHeaders: http.Header{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: http.Header{
				"Access-Control-Allow-Origin":  {"https://foo.bar.org"},
				"Access-Control-Max-Age":       {"100"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
			},
			reqHost: "test.localhost",
			method:  http.MethodOptions,
		},
		{
			desc: "preflight Options request with no cors configured",
			requestHeaders: http.Header{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: http.Header{
				"X-Custom-Response-Header": {"True"},
			},
			reqHost: "test2.localhost",
			method:  http.MethodOptions,
		},
		{
			desc: "preflight Get request with no cors configured",
			requestHeaders: http.Header{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: http.Header{
				"X-Custom-Response-Header": {"True"},
			},
			reqHost: "test2.localhost",
			method:  http.MethodGet,
		},
	}

	for _, test := range testCase {
		req, err := http.NewRequest(test.method, "http://127.0.0.1:8000/", nil)
		require.NoError(s.T(), err)
		req.Host = test.reqHost
		req.Header = test.requestHeaders

		err = try.Request(req, 500*time.Millisecond, try.HasHeaderStruct(test.expected))
		require.NoError(s.T(), err)
	}
}

func (s *HeadersSuite) TestSecureHeadersResponses() {
	file := s.adaptFile("fixtures/headers/secure.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	backend := startTestServer("9000", http.StatusOK, "")
	defer backend.Close()

	err := try.GetRequest(backend.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	testCase := []struct {
		desc            string
		expected        http.Header
		reqHost         string
		internalReqHost string
	}{
		{
			desc: "Permissions-Policy Set",
			expected: http.Header{
				"Permissions-Policy": {"microphone=(),"},
			},
			reqHost:         "test.localhost",
			internalReqHost: "internal.localhost",
		},
	}

	for _, test := range testCase {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
		require.NoError(s.T(), err)
		req.Host = test.reqHost

		err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasHeaderStruct(test.expected))
		require.NoError(s.T(), err)

		req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/api/rawdata", nil)
		require.NoError(s.T(), err)
		req.Host = test.internalReqHost

		err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasHeaderStruct(test.expected))
		require.NoError(s.T(), err)
	}
}

func (s *HeadersSuite) TestMultipleSecureHeadersResponses() {
	file := s.adaptFile("fixtures/headers/secure_multiple.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	backend := startTestServer("9000", http.StatusOK, "")
	defer backend.Close()

	err := try.GetRequest(backend.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	testCase := []struct {
		desc     string
		expected http.Header
		reqHost  string
	}{
		{
			desc: "Multiple Secure Headers Set",
			expected: http.Header{
				"X-Frame-Options":        {"DENY"},
				"X-Content-Type-Options": {"nosniff"},
			},
			reqHost: "test.localhost",
		},
	}

	for _, test := range testCase {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
		require.NoError(s.T(), err)
		req.Host = test.reqHost

		err = try.Request(req, 500*time.Millisecond, try.HasHeaderStruct(test.expected))
		require.NoError(s.T(), err)
	}
}
