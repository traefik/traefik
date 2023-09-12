package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-check/check"
	"github.com/kvtools/consul"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/traefik/traefik/v2/integration/try"
	"github.com/traefik/traefik/v2/pkg/api"
	checker "github.com/vdemeester/shakers"
)

// Consul test suites.
type ConsulSuite struct {
	BaseSuite
	kvClient  store.Store
	consulURL string
}

func (s *ConsulSuite) resetStore(c *check.C) {
	err := s.kvClient.DeleteTree(context.Background(), "traefik")
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		c.Fatal(err)
	}
}

func (s *ConsulSuite) setupStore(c *check.C) {
	s.createComposeProject(c, "consul")
	s.composeUp(c)

	consulAddr := net.JoinHostPort(s.getComposeServiceIP(c, "consul"), "8500")
	s.consulURL = fmt.Sprintf("http://%s", consulAddr)

	kv, err := valkeyrie.NewStore(
		context.Background(),
		consul.StoreName,
		[]string{consulAddr},
		&consul.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		c.Fatal("Cannot create store consul")
	}
	s.kvClient = kv

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(kv, "test"))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) TestSimpleConfiguration(c *check.C) {
	s.setupStore(c)

	file := s.adaptFile(c, "fixtures/consul/simple.toml", struct{ ConsulAddress string }{s.consulURL})
	defer os.Remove(file)

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
		"traefik/http/middlewares/striper/stripPrefix/forceSlash": "true",
	}

	for k, v := range data {
		err := s.kvClient.Put(context.Background(), k, []byte(v), nil)
		c.Assert(err, checker.IsNil)
	}

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.BodyContains(`"striper@consul":`, `"compressor@consul":`, `"srvcA@consul":`, `"srvcB@consul":`),
	)
	c.Assert(err, checker.IsNil)

	resp, err := http.Get("http://127.0.0.1:8080/api/rawdata")
	c.Assert(err, checker.IsNil)

	var obtained api.RunTimeRepresentation
	err = json.NewDecoder(resp.Body).Decode(&obtained)
	c.Assert(err, checker.IsNil)
	got, err := json.MarshalIndent(obtained, "", "  ")
	c.Assert(err, checker.IsNil)

	expectedJSON := filepath.FromSlash("testdata/rawdata-consul.json")

	if *updateExpected {
		err = os.WriteFile(expectedJSON, got, 0o666)
		c.Assert(err, checker.IsNil)
	}

	expected, err := os.ReadFile(expectedJSON)
	c.Assert(err, checker.IsNil)

	if !bytes.Equal(expected, got) {
		diff := difflib.UnifiedDiff{
			FromFile: "Expected",
			A:        difflib.SplitLines(string(expected)),
			ToFile:   "Got",
			B:        difflib.SplitLines(string(got)),
			Context:  3,
		}

		text, err := difflib.GetUnifiedDiffString(diff)
		c.Assert(err, checker.IsNil)
		c.Error(text)
	}
}

func (s *ConsulSuite) assertWhoami(c *check.C, host string, expectedStatusCode int) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	if err != nil {
		c.Fatal(err)
	}
	req.Host = host

	resp, err := try.ResponseUntilStatusCode(req, 15*time.Second, expectedStatusCode)
	resp.Body.Close()
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) TestDeleteRootKey(c *check.C) {
	// This test case reproduce the issue: https://github.com/traefik/traefik/issues/8092
	s.setupStore(c)
	s.resetStore(c)

	file := s.adaptFile(c, "fixtures/consul/simple.toml", struct{ ConsulAddress string }{s.consulURL})
	defer os.Remove(file)

	ctx := context.Background()
	svcaddr := net.JoinHostPort(s.getComposeServiceIP(c, "whoami"), "80")

	data := map[string]string{
		"traefik/http/routers/Router0/entryPoints/0": "web",
		"traefik/http/routers/Router0/rule":          "Host(`kv1.localhost`)",
		"traefik/http/routers/Router0/service":       "simplesvc0",

		"traefik/http/routers/Router1/entryPoints/0": "web",
		"traefik/http/routers/Router1/rule":          "Host(`kv2.localhost`)",
		"traefik/http/routers/Router1/service":       "simplesvc1",

		"traefik/http/services/simplesvc0/loadBalancer/servers/0/url": "http://" + svcaddr,
		"traefik/http/services/simplesvc1/loadBalancer/servers/0/url": "http://" + svcaddr,
	}

	for k, v := range data {
		err := s.kvClient.Put(ctx, k, []byte(v), nil)
		c.Assert(err, checker.IsNil)
	}

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.BodyContains(`"Router0@consul":`, `"Router1@consul":`, `"simplesvc0@consul":`, `"simplesvc1@consul":`),
	)
	c.Assert(err, checker.IsNil)
	s.assertWhoami(c, "kv1.localhost", http.StatusOK)
	s.assertWhoami(c, "kv2.localhost", http.StatusOK)

	// delete router1
	err = s.kvClient.DeleteTree(ctx, "traefik/http/routers/Router1")
	c.Assert(err, checker.IsNil)
	s.assertWhoami(c, "kv1.localhost", http.StatusOK)
	s.assertWhoami(c, "kv2.localhost", http.StatusNotFound)

	// delete simple services and router0
	err = s.kvClient.DeleteTree(ctx, "traefik")
	c.Assert(err, checker.IsNil)
	s.assertWhoami(c, "kv1.localhost", http.StatusNotFound)
	s.assertWhoami(c, "kv2.localhost", http.StatusNotFound)
}
