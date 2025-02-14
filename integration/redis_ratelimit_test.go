package integration

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type RedisRateLimitSuite struct {
	BaseSuite
	ServerIP       string
	redisEndpoints []string
}

func TestRedisRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RedisRateLimitSuite))
}

func (s *RedisRateLimitSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("ratelimit_redis")
	s.composeUp()

	s.redisEndpoints = []string{}
	s.redisEndpoints = append(s.redisEndpoints, net.JoinHostPort(s.getComposeServiceIP("redis"), "6379"))

	s.ServerIP = s.getComposeServiceIP("whoami1")
}

func (s *RedisRateLimitSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *RedisRateLimitSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/ratelimit/simple_redis.toml", struct {
		Server1      string
		RedisAddress string
	}{
		Server1:      s.ServerIP,
		RedisAddress: strings.Join(s.redisEndpoints, ","),
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
