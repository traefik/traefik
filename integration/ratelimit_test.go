package integration

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type RateLimitSuite struct {
	BaseSuite
	ServerIP      string
	RedisEndpoint string
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}

func (s *RateLimitSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("ratelimit")
	s.composeUp()

	s.ServerIP = s.getComposeServiceIP("whoami1")
	s.RedisEndpoint = net.JoinHostPort(s.getComposeServiceIP("redis"), "6379")
}

func (s *RateLimitSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *RateLimitSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/ratelimit/simple.toml", struct {
		Server1 string
	}{s.ServerIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("ratelimit"))
	require.NoError(s.T(), err)

	start := time.Now()
	count := 0
	for {
		err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
		require.NoError(s.T(), err)
		count++
		if count > 100 {
			break
		}
	}
	stop := time.Now()
	elapsed := stop.Sub(start)
	if elapsed < time.Second*99/100 {
		s.T().Fatalf("requests throughput was too fast wrt to rate limiting: 100 requests in %v", elapsed)
	}
}

func (s *RateLimitSuite) TestRedisRateLimitSimpleConfiguration() {
	file := s.adaptFile("fixtures/ratelimit/simple_redis.toml", struct {
		Server1       string
		RedisEndpoint string
	}{
		Server1:       s.ServerIP,
		RedisEndpoint: s.RedisEndpoint,
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("ratelimit", "redis"))
	require.NoError(s.T(), err)

	start := time.Now()
	count := 0
	for {
		err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
		require.NoError(s.T(), err)
		count++
		if count > 100 {
			break
		}
	}
	stop := time.Now()
	elapsed := stop.Sub(start)
	if elapsed < time.Second*99/100 {
		s.T().Fatalf("requests throughput was too fast wrt to rate limiting: 100 requests in %v", elapsed)
	}
}
