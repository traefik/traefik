package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
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
	"github.com/traefik/traefik/v2/integration/try"
	"github.com/traefik/traefik/v2/pkg/api"
)

// Redis test suites.
type RedisSuite struct {
	BaseSuite
	kvClient       store.Store
	redisEndpoints []string
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}

func (s *RedisSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("redis")
	s.composeUp()

	s.redisEndpoints = []string{}
	s.redisEndpoints = append(s.redisEndpoints, net.JoinHostPort(s.getComposeServiceIP("redis"), "6379"))

	kv, err := valkeyrie.NewStore(
		context.Background(),
		redis.StoreName,
		s.redisEndpoints,
		&redis.Config{},
	)
	if err != nil {
		s.T().Fatal("Cannot create store redis: ", err)
	}
	s.kvClient = kv

	// wait for redis
	err = try.Do(60*time.Second, try.KVExists(kv, "test"))
	require.NoError(s.T(), err)
}

func (s *RedisSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()

	for _, filename := range []string{"sentinel1.conf", "sentinel2.conf", "sentinel3.conf"} {
		err := os.Remove(filepath.Join(".", "resources", "compose", "config", filename))
		require.NotErrorIs(s.T(), err, fs.ErrNotExist)
	}
}

func (s *RedisSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/redis/simple.toml", struct{ RedisAddress string }{
		RedisAddress: strings.Join(s.redisEndpoints, ","),
	})
	defer os.Remove(file)

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
		"traefik/http/middlewares/striper/stripPrefix/forceSlash": "true",
	}

	for k, v := range data {
		err := s.kvClient.Put(context.Background(), k, []byte(v), nil)
		require.NoError(s.T(), err)
	}

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display()
	err := cmd.Start()
	require.NoError(s.T(), err)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
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
		require.NoError(s.T(), err, text)
	}
}

func (s *RedisSuite) setupSentinelStore() {
	s.setupSentinelConfiguration([]string{"26379", "36379", "46379"})

	s.createComposeProject("redis_sentinel")
	s.composeUp("redis_sentinel")

	s.redisEndpoints = []string{
		net.JoinHostPort(s.getComposeServiceIP("sentinel1"), "26379"),
		net.JoinHostPort(s.getComposeServiceIP("sentinel2"), "36379"),
		net.JoinHostPort(s.getComposeServiceIP("sentinel3"), "46379"),
	}

	kv, err := valkeyrie.NewStore(
		context.Background(),
		redis.StoreName,
		s.redisEndpoints,
		&redis.Config{
			Sentinel: &redis.Sentinel{
				MasterName: "mymaster",
			},
		},
	)
	require.NoError(s.T(), err)
	s.kvClient = kv

	// wait for redis
	err = try.Do(60*time.Second, try.KVExists(kv, "test"))
	require.NoError(s.T(), err)
}

func (s *RedisSuite) setupSentinelConfiguration(ports []string) {
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

		model := structs.Map(templateValue)
		model["SelfFilename"] = tmpFile.Name()

		err = tmpl.ExecuteTemplate(tmpFile, prefix, model)
		require.NoError(s.T(), err)

		err = tmpFile.Sync()
		require.NoError(s.T(), err)
	}
}

func (s *RedisSuite) TestSentinelConfiguration() {
	s.setupSentinelStore()

	file := s.adaptFile("fixtures/redis/sentinel.toml", struct{ RedisAddress string }{
		RedisAddress: strings.Join(s.redisEndpoints, `","`),
	})
	defer os.Remove(file)

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
		"traefik/http/middlewares/striper/stripPrefix/forceSlash": "true",
	}

	for k, v := range data {
		err := s.kvClient.Put(context.Background(), k, []byte(v), nil)
		require.NoError(s.T(), err)
	}

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display()
	err := cmd.Start()
	require.NoError(s.T(), err)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
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
