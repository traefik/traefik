package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v2/integration/try"
)

type TracingSuite struct {
	BaseSuite
	whoamiIP       string
	whoamiPort     int
	tracerZipkinIP string
	tracerJaegerIP string
}

func TestTracingSuite(t *testing.T) {
	suite.Run(t, new(TracingSuite))
}

type TracingTemplate struct {
	WhoamiIP               string
	WhoamiPort             int
	IP                     string
	TraceContextHeaderName string
}

func (s *TracingSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("tracing")
	s.composeUp()

	s.whoamiIP = s.getComposeServiceIP("whoami")
	s.whoamiPort = 80
}

func (s *TracingSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TracingSuite) startZipkin() {
	s.composeUp("zipkin")
	s.tracerZipkinIP = s.getComposeServiceIP("zipkin")

	// Wait for Zipkin to turn ready.
	err := try.GetRequest("http://"+s.tracerZipkinIP+":9411/api/v2/services", 20*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestZipkinRateLimit() {
	s.startZipkin()

	file := s.adaptFile("fixtures/tracing/simple-zipkin.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.tracerZipkinIP,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	require.NoError(s.T(), err)

	// sleep for 4 seconds to be certain the configured time period has elapsed
	// then test another request and verify a 200 status code
	time.Sleep(4 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// continue requests at 3 second intervals to test the other rate limit time period
	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerZipkinIP+":9411/api/v2/spans?serviceName=tracing", 20*time.Second, try.BodyContains("forward service1/router1@file", "ratelimit-1@file"))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestZipkinRetry() {
	s.startZipkin()

	file := s.adaptFile("fixtures/tracing/simple-zipkin.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: 81,
		IP:         s.tracerZipkinIP,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/retry", 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerZipkinIP+":9411/api/v2/spans?serviceName=tracing", 20*time.Second, try.BodyContains("forward service2/router2@file", "retry@file"))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestZipkinAuth() {
	s.startZipkin()

	file := s.adaptFile("fixtures/tracing/simple-zipkin.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.tracerZipkinIP,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerZipkinIP+":9411/api/v2/spans?serviceName=tracing", 20*time.Second, try.BodyContains("entrypoint web", "basic-auth@file"))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) startJaeger() {
	s.composeUp("jaeger", "whoami")
	s.tracerJaegerIP = s.getComposeServiceIP("jaeger")

	// Wait for Jaeger to turn ready.
	err := try.GetRequest("http://"+s.tracerJaegerIP+":16686/api/services", 20*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestJaegerRateLimit() {
	s.startJaeger()
	// defer s.composeStop(c, "jaeger")

	file := s.adaptFile("fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             s.whoamiPort,
		IP:                     s.tracerJaegerIP,
		TraceContextHeaderName: "uber-trace-id",
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	require.NoError(s.T(), err)

	// sleep for 4 seconds to be certain the configured time period has elapsed
	// then test another request and verify a 200 status code
	time.Sleep(4 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// continue requests at 3 second intervals to test the other rate limit time period
	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerJaegerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("forward service1/router1@file", "ratelimit-1@file"))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestJaegerRetry() {
	s.startJaeger()
	// defer s.composeStop(c, "jaeger")

	file := s.adaptFile("fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             81,
		IP:                     s.tracerJaegerIP,
		TraceContextHeaderName: "uber-trace-id",
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/retry", 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerJaegerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("forward service2/router2@file", "retry@file"))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestJaegerAuth() {
	s.startJaeger()
	// defer s.composeStop(c, "jaeger")

	file := s.adaptFile("fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             s.whoamiPort,
		IP:                     s.tracerJaegerIP,
		TraceContextHeaderName: "uber-trace-id",
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerJaegerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("EntryPoint web", "basic-auth@file"))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestJaegerCustomHeader() {
	s.startJaeger()
	// defer s.composeStop(c, "jaeger")

	file := s.adaptFile("fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             s.whoamiPort,
		IP:                     s.tracerJaegerIP,
		TraceContextHeaderName: "powpow",
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerJaegerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("EntryPoint web", "basic-auth@file"))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TestJaegerAuthCollector() {
	s.startJaeger()
	// defer s.composeStop(c, "jaeger")

	file := s.adaptFile("fixtures/tracing/simple-jaeger-collector.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.tracerJaegerIP,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+s.tracerJaegerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("EntryPoint web", "basic-auth@file"))
	require.NoError(s.T(), err)
}
