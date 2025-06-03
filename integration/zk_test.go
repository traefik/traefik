package integration

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"
	"github.com/kvtools/zookeeper"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/api"
)

// Zk test suites.
type ZookeeperSuite struct {
	BaseSuite
	kvClient      store.Store
	zookeeperAddr string
}

func TestZookeeperSuite(t *testing.T) {
	suite.Run(t, new(ZookeeperSuite))
}

func (s *ZookeeperSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("zookeeper")
	s.composeUp()

	s.zookeeperAddr = net.JoinHostPort(s.getComposeServiceIP("zookeeper"), "2181")

	var err error
	s.kvClient, err = valkeyrie.NewStore(
		s.T().Context(),
		zookeeper.StoreName,
		[]string{s.zookeeperAddr},
		&zookeeper.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	require.NoError(s.T(), err, "Cannot create store zookeeper")

	// wait for zk
	err = try.Do(60*time.Second, try.KVExists(s.kvClient, "test"))
	require.NoError(s.T(), err)
}

func (s *ZookeeperSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *ZookeeperSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/zookeeper/simple.toml", struct{ ZkAddress string }{s.zookeeperAddr})

	data := map[string]string{
		"traefik/http/routers/Router0/entryPoints/0": "web",
		"traefik/http/routers/Router0/middlewares/0": "compressor",
		"traefik/http/routers/Router0/middlewares/1": "striper",
		"traefik/http/routers/Router0/service":       "simplesvc",
		"traefik/http/routers/Router0/rule":          "Host(`kv1.localhost`)",
		"traefik/http/routers/Router0/priority":      "42",
		"traefik/http/routers/Router0/tls":           "",

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

		"traefik/http/middlewares/compressor/compress":            "",
		"traefik/http/middlewares/striper/stripPrefix/prefixes/0": "foo",
		"traefik/http/middlewares/striper/stripPrefix/prefixes/1": "bar",
	}

	for k, v := range data {
		err := s.kvClient.Put(s.T().Context(), k, []byte(v), nil)
		require.NoError(s.T(), err)
	}

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second,
		try.BodyContains(`"striper@zookeeper":`, `"compressor@zookeeper":`, `"srvcA@zookeeper":`, `"srvcB@zookeeper":`),
	)
	require.NoError(s.T(), err)

	resp, err := http.Get("http://127.0.0.1:8080/api/rawdata")
	require.NoError(s.T(), err)

	var obtained api.RunTimeRepresentation
	err = json.NewDecoder(resp.Body).Decode(&obtained)
	require.NoError(s.T(), err)
	got, err := json.MarshalIndent(obtained, "", "  ")
	require.NoError(s.T(), err)

	expectedJSON := filepath.FromSlash("testdata/rawdata-zk.json")

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
