package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/fatih/structs"
	"github.com/kvtools/redis"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/api"
)

// Redis test suites.
type RedisSentinelSuite struct {
	BaseSuite
	kvClient       store.Store
	redisEndpoints []string
}

func TestRedisSentinelSuite(t *testing.T) {
	suite.Run(t, new(RedisSentinelSuite))
}

func (s *RedisSentinelSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.setupSentinelConfiguration([]string{"26379", "26379", "26379"})

	s.createComposeProject("redis_sentinel")
	s.composeUp()

	s.redisEndpoints = []string{
		net.JoinHostPort(s.getComposeServiceIP("sentinel1"), "26379"),
		net.JoinHostPort(s.getComposeServiceIP("sentinel2"), "26379"),
		net.JoinHostPort(s.getComposeServiceIP("sentinel3"), "26379"),
	}
	kv, err := valkeyrie.NewStore(
		s.T().Context(),
		redis.StoreName,
		s.redisEndpoints,
		&redis.Config{
			Sentinel: &redis.Sentinel{
				MasterName: "mymaster",
			},
		},
	)
	require.NoError(s.T(), err, "Cannot create store redis")
	s.kvClient = kv

	// wait for redis
	err = try.Do(60*time.Second, try.KVExists(kv, "test"))
	require.NoError(s.T(), err)
}

func (s *RedisSentinelSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()

	for _, filename := range []string{"sentinel1.conf", "sentinel2.conf", "sentinel3.conf"} {
		_ = os.Remove(filepath.Join(".", "resources", "compose", "config", filename))
	}
}

func (s *RedisSentinelSuite) setupSentinelConfiguration(ports []string) {
	for i, port := range ports {
		templateValue := struct{ SentinelPort string }{SentinelPort: port}

		// Load file
		templateFile := "resources/compose/config/sentinel_template.conf"
		tmpl, err := template.ParseFiles(templateFile)
		require.NoError(s.T(), err)

		folder, prefix := filepath.Split(templateFile)

		fileName := fmt.Sprintf("%s/sentinel%d.conf", folder, i+1)
		tmpFile, err := os.Create(fileName)
		require.NoError(s.T(), err)
		defer tmpFile.Close()

		err = tmpFile.Chmod(0o666)
		require.NoError(s.T(), err)

		model := structs.Map(templateValue)
		model["SelfFilename"] = tmpFile.Name()

		err = tmpl.ExecuteTemplate(tmpFile, prefix, model)
		require.NoError(s.T(), err)

		err = tmpFile.Sync()
		require.NoError(s.T(), err)
	}
}

func (s *RedisSentinelSuite) TestSentinelConfiguration() {
	file := s.adaptFile("fixtures/redis/sentinel.toml", struct{ RedisAddress string }{
		RedisAddress: strings.Join(s.redisEndpoints, `","`),
	})

	data := map[string]string{
		"traefik/http/routers/Router0/entryPoints/0": "web",
		"traefik/http/routers/Router0/middlewares/0": "compressor",
		"traefik/http/routers/Router0/middlewares/1": "striper",
		"traefik/http/routers/Router0/service":       "simplesvc",
		"traefik/http/routers/Router0/rule":          "Host(`kv1.localhost`)",
		"traefik/http/routers/Router0/priority":      "42",
		"traefik/http/routers/Router0/tls":           "true",

		"traefik/http/routers/Router1/rule":                 "Host(`kv2.localhost`)",
		"traefik/http/routers/Router1/priority":             "42",
		"traefik/http/routers/Router1/tls/domains/0/main":   "aaa.localhost",
		"traefik/http/routers/Router1/tls/domains/0/sans/0": "aaa.aaa.localhost",
		"traefik/http/routers/Router1/tls/domains/0/sans/1": "bbb.aaa.localhost",
		"traefik/http/routers/Router1/tls/domains/1/main":   "bbb.localhost",
		"traefik/http/routers/Router1/tls/domains/1/sans/0": "aaa.bbb.localhost",
		"traefik/http/routers/Router1/tls/domains/1/sans/1": "bbb.bbb.localhost",
		"traefik/http/routers/Router1/entryPoints/0":        "web",
		"traefik/http/routers/Router1/service":              "simplesvc",

		"traefik/http/services/simplesvc/loadBalancer/servers/0/url": "http://10.0.1.1:8888",
		"traefik/http/services/simplesvc/loadBalancer/servers/1/url": "http://10.0.1.1:8889",

		"traefik/http/services/srvcA/loadBalancer/servers/0/url": "http://10.0.1.2:8888",
		"traefik/http/services/srvcA/loadBalancer/servers/1/url": "http://10.0.1.2:8889",

		"traefik/http/services/srvcB/loadBalancer/servers/0/url": "http://10.0.1.3:8888",
		"traefik/http/services/srvcB/loadBalancer/servers/1/url": "http://10.0.1.3:8889",

		"traefik/http/services/mirror/mirroring/service":           "simplesvc",
		"traefik/http/services/mirror/mirroring/mirrors/0/name":    "srvcA",
		"traefik/http/services/mirror/mirroring/mirrors/0/percent": "42",
		"traefik/http/services/mirror/mirroring/mirrors/1/name":    "srvcB",
		"traefik/http/services/mirror/mirroring/mirrors/1/percent": "42",

		"traefik/http/services/Service03/weighted/services/0/name":   "srvcA",
		"traefik/http/services/Service03/weighted/services/0/weight": "42",
		"traefik/http/services/Service03/weighted/services/1/name":   "srvcB",
		"traefik/http/services/Service03/weighted/services/1/weight": "42",

		"traefik/http/middlewares/compressor/compress":            "true",
		"traefik/http/middlewares/striper/stripPrefix/prefixes/0": "foo",
		"traefik/http/middlewares/striper/stripPrefix/prefixes/1": "bar",
	}

	for k, v := range data {
		err := s.kvClient.Put(s.T().Context(), k, []byte(v), nil)
		require.NoError(s.T(), err)
	}

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.BodyContains(`"striper@redis":`, `"compressor@redis":`, `"srvcA@redis":`, `"srvcB@redis":`),
	)
	require.NoError(s.T(), err)

	resp, err := http.Get("http://127.0.0.1:8080/api/rawdata")
	require.NoError(s.T(), err)

	var obtained api.RunTimeRepresentation
	err = json.NewDecoder(resp.Body).Decode(&obtained)
	require.NoError(s.T(), err)
	got, err := json.MarshalIndent(obtained, "", "  ")
	require.NoError(s.T(), err)

	expectedJSON := filepath.FromSlash("testdata/rawdata-redis.json")

	if *updateExpected {
		err = os.WriteFile(expectedJSON, got, 0o666)
		require.NoError(s.T(), err)
	}

	expected, err := os.ReadFile(expectedJSON)
	require.NoError(s.T(), err)

	if !bytes.Equal(expected, got) {
		diff := difflib.UnifiedDiff{
			FromFile: "Expected",
			A:        difflib.SplitLines(string(expected)),
			ToFile:   "Got",
			B:        difflib.SplitLines(string(got)),
			Context:  3,
		}

		text, err := difflib.GetUnifiedDiffString(diff)
		require.NoError(s.T(), err)
		log.Info().Msg(text)
	}
}
