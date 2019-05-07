package integration

import (
	"net/http"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// Headers test suites
type HeadersSuite struct{ BaseSuite }

func (s *HeadersSuite) TestSimpleConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/headers/basic.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *HeadersSuite) TestCorsResponses(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/headers/cors.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	backend := startTestServer("9000", http.StatusOK)
	defer backend.Close()

	err = try.GetRequest(backend.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	testCase := []struct {
		desc           string
		requestHeaders http.Header
		expected       http.Header
	}{
		{
			desc: "simple access control allow origin",
			requestHeaders: http.Header{
				"Origin": {"http://test.localhost"},
			},
			expected: http.Header{
				"Access-Control-Allow-Origin": {"http://test.localhost"},
				"Vary":                        {"Origin"},
			},
		},
	}

	for _, test := range testCase {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
		c.Assert(err, checker.IsNil)
		req.Host = "test.localhost"
		req.Header = test.requestHeaders

		err = try.Request(req, 500*time.Millisecond, try.HasHeaderStruct(test.expected))
		c.Assert(err, checker.IsNil)
	}
}

func (s *HeadersSuite) TestCorsPreflightResponses(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/headers/cors.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	backend := startTestServer("9000", http.StatusOK)
	defer backend.Close()

	err = try.GetRequest(backend.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	testCase := []struct {
		desc           string
		requestHeaders http.Header
		expected       http.Header
	}{
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
		},
	}

	for _, test := range testCase {
		req, err := http.NewRequest(http.MethodOptions, "http://127.0.0.1:8000/", nil)
		c.Assert(err, checker.IsNil)
		req.Host = "test.localhost"
		req.Header = test.requestHeaders

		err = try.Request(req, 500*time.Millisecond, try.HasHeaderStruct(test.expected))
		c.Assert(err, checker.IsNil)
	}
}
