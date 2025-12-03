package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/baqupio/baqup/v3/integration/try"
	"github.com/baqupio/baqup/v3/pkg/api"
	"github.com/kvtools/consul"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Consul test suites.
type ConsulSuite struct {
	BaseSuite
	kvClient  store.Store
	consulURL string
}

func TestConsulSuite(t *testing.T) {
	suite.Run(t, new(ConsulSuite))
}

func (s *ConsulSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.createComposeProject("consul")
	s.composeUp()

	consulAddr := net.JoinHostPort(s.getComposeServiceIP("consul"), "8500")
	s.consulURL = fmt.Sprintf("http://%s", consulAddr)

	kv, err := valkeyrie.NewStore(
		s.T().Context(),
		consul.StoreName,
		[]string{consulAddr},
		&consul.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	require.NoError(s.T(), err, "Cannot create store consul")
	s.kvClient = kv

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(kv, "test"))
	require.NoError(s.T(), err)
}

func (s *ConsulSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *ConsulSuite) TearDownTest() {
	err := s.kvClient.DeleteTree(s.T().Context(), "baqup")
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		require.ErrorIs(s.T(), err, store.ErrKeyNotFound)
	}
}

func (s *ConsulSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/consul/simple.toml", struct{ ConsulAddress string }{s.consulURL})

	data := map[string]string{
		"baqup/http/routers/Router0/entryPoints/0": "web",
		"baqup/http/routers/Router0/middlewares/0": "compressor",
		"baqup/http/routers/Router0/middlewares/1": "striper",
		"baqup/http/routers/Router0/service":       "simplesvc",
		"baqup/http/routers/Router0/rule":          "Host(`kv1.localhost`)",
		"baqup/http/routers/Router0/priority":      "42",
		"baqup/http/routers/Router0/tls":           "",

		"baqup/http/routers/Router1/rule":                 "Host(`kv2.localhost`)",
		"baqup/http/routers/Router1/priority":             "42",
		"baqup/http/routers/Router1/tls/domains/0/main":   "aaa.localhost",
		"baqup/http/routers/Router1/tls/domains/0/sans/0": "aaa.aaa.localhost",
		"baqup/http/routers/Router1/tls/domains/0/sans/1": "bbb.aaa.localhost",
		"baqup/http/routers/Router1/tls/domains/1/main":   "bbb.localhost",
		"baqup/http/routers/Router1/tls/domains/1/sans/0": "aaa.bbb.localhost",
		"baqup/http/routers/Router1/tls/domains/1/sans/1": "bbb.bbb.localhost",
		"baqup/http/routers/Router1/entryPoints/0":        "web",
		"baqup/http/routers/Router1/service":              "simplesvc",

		"baqup/http/services/simplesvc/loadBalancer/servers/0/url": "http://10.0.1.1:8888",
		"baqup/http/services/simplesvc/loadBalancer/servers/1/url": "http://10.0.1.1:8889",

		"baqup/http/services/srvcA/loadBalancer/servers/0/url": "http://10.0.1.2:8888",
		"baqup/http/services/srvcA/loadBalancer/servers/1/url": "http://10.0.1.2:8889",

		"baqup/http/services/srvcB/loadBalancer/servers/0/url": "http://10.0.1.3:8888",
		"baqup/http/services/srvcB/loadBalancer/servers/1/url": "http://10.0.1.3:8889",

		"baqup/http/services/mirror/mirroring/service":           "simplesvc",
		"baqup/http/services/mirror/mirroring/mirrors/0/name":    "srvcA",
		"baqup/http/services/mirror/mirroring/mirrors/0/percent": "42",
		"baqup/http/services/mirror/mirroring/mirrors/1/name":    "srvcB",
		"baqup/http/services/mirror/mirroring/mirrors/1/percent": "42",

		"baqup/http/services/Service03/weighted/services/0/name":   "srvcA",
		"baqup/http/services/Service03/weighted/services/0/weight": "42",
		"baqup/http/services/Service03/weighted/services/1/name":   "srvcB",
		"baqup/http/services/Service03/weighted/services/1/weight": "42",

		"baqup/http/middlewares/compressor/compress":            "",
		"baqup/http/middlewares/striper/stripPrefix/prefixes/0": "foo",
		"baqup/http/middlewares/striper/stripPrefix/prefixes/1": "bar",
	}

	for k, v := range data {
		err := s.kvClient.Put(s.T().Context(), k, []byte(v), nil)
		require.NoError(s.T(), err)
	}

	s.baqupCmd(withConfigFile(file))

	// wait for baqup
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.BodyContains(`"striper@consul":`, `"compressor@consul":`, `"srvcA@consul":`, `"srvcB@consul":`),
	)
	require.NoError(s.T(), err)

	resp, err := http.Get("http://127.0.0.1:8080/api/rawdata")
	require.NoError(s.T(), err)

	var obtained api.RunTimeRepresentation
	err = json.NewDecoder(resp.Body).Decode(&obtained)
	require.NoError(s.T(), err)
	got, err := json.MarshalIndent(obtained, "", "  ")
	require.NoError(s.T(), err)

	expectedJSON := filepath.FromSlash("testdata/rawdata-consul.json")

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

func (s *ConsulSuite) assertWhoami(host string, expectedStatusCode int) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	require.NoError(s.T(), err)
	req.Host = host

	resp, err := try.ResponseUntilStatusCode(req, 15*time.Second, expectedStatusCode)
	require.NoError(s.T(), err)
	resp.Body.Close()
}

func (s *ConsulSuite) TestDeleteRootKey() {
	// This test case reproduce the issue: https://github.com/baqupio/baqup/issues/8092

	file := s.adaptFile("fixtures/consul/simple.toml", struct{ ConsulAddress string }{s.consulURL})

	ctx := s.T().Context()
	svcaddr := net.JoinHostPort(s.getComposeServiceIP("whoami"), "80")

	data := map[string]string{
		"baqup/http/routers/Router0/entryPoints/0": "web",
		"baqup/http/routers/Router0/rule":          "Host(`kv1.localhost`)",
		"baqup/http/routers/Router0/service":       "simplesvc0",

		"baqup/http/routers/Router1/entryPoints/0": "web",
		"baqup/http/routers/Router1/rule":          "Host(`kv2.localhost`)",
		"baqup/http/routers/Router1/service":       "simplesvc1",

		"baqup/http/services/simplesvc0/loadBalancer/servers/0/url": "http://" + svcaddr,
		"baqup/http/services/simplesvc1/loadBalancer/servers/0/url": "http://" + svcaddr,
	}

	for k, v := range data {
		err := s.kvClient.Put(ctx, k, []byte(v), nil)
		require.NoError(s.T(), err)
	}

	s.baqupCmd(withConfigFile(file))

	// wait for baqup
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.BodyContains(`"Router0@consul":`, `"Router1@consul":`, `"simplesvc0@consul":`, `"simplesvc1@consul":`),
	)
	require.NoError(s.T(), err)
	s.assertWhoami("kv1.localhost", http.StatusOK)
	s.assertWhoami("kv2.localhost", http.StatusOK)

	// delete router1
	err = s.kvClient.DeleteTree(ctx, "baqup/http/routers/Router1")
	require.NoError(s.T(), err)
	s.assertWhoami("kv1.localhost", http.StatusOK)
	s.assertWhoami("kv2.localhost", http.StatusNotFound)

	// delete simple services and router0
	err = s.kvClient.DeleteTree(ctx, "baqup")
	require.NoError(s.T(), err)
	s.assertWhoami("kv1.localhost", http.StatusNotFound)
	s.assertWhoami("kv2.localhost", http.StatusNotFound)
}
