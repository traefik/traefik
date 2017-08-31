package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/containous/staert"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// Consul test suites (using libcompose)
type ConsulSuite struct {
	BaseSuite
	kv store.Store
}

func (s *ConsulSuite) setupConsul(c *check.C) {
	s.createComposeProject(c, "consul")
	s.composeProject.Start(c)

	consul.Register()
	kv, err := libkv.NewStore(
		store.CONSUL,
		[]string{s.composeProject.Container(c, "consul").NetworkSettings.IPAddress + ":8500"},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		c.Fatal("Cannot create store consul")
	}
	s.kv = kv

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(kv, "test"))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) setupConsulTLS(c *check.C) {
	s.createComposeProject(c, "consul_tls")
	s.composeProject.Start(c)

	consul.Register()
	clientTLS := &types.ClientTLS{
		CA:                 "resources/tls/ca.cert",
		Cert:               "resources/tls/consul.cert",
		Key:                "resources/tls/consul.key",
		InsecureSkipVerify: true,
	}
	TLSConfig, err := clientTLS.CreateTLSConfig()
	c.Assert(err, checker.IsNil)

	kv, err := libkv.NewStore(
		store.CONSUL,
		[]string{s.composeProject.Container(c, "consul").NetworkSettings.IPAddress + ":8585"},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
			TLS:               TLSConfig,
		},
	)

	if err != nil {
		c.Fatal("Cannot create store consul")
	}
	s.kv = kv

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(kv, "test"))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) TearDownTest(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *ConsulSuite) TearDownSuite(c *check.C) {}

func (s *ConsulSuite) TestSimpleConfiguration(c *check.C) {
	s.setupConsul(c)
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/consul/simple.toml", struct{ ConsulHost string }{consulHost})
	defer os.Remove(file)

	cmd, _ := s.cmdTraefik(withConfigFile(file))
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) TestNominalConfiguration(c *check.C) {
	s.setupConsul(c)
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/consul/simple.toml", struct{ ConsulHost string }{consulHost})
	defer os.Remove(file)

	cmd, _ := s.cmdTraefik(withConfigFile(file))
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami1IP := s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress
	whoami2IP := s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress
	whoami3IP := s.composeProject.Container(c, "whoami3").NetworkSettings.IPAddress
	whoami4IP := s.composeProject.Container(c, "whoami4").NetworkSettings.IPAddress

	backend1 := map[string]string{
		"traefik/backends/backend1/circuitbreaker/expression": "NetworkErrorRatio() > 0.5",
		"traefik/backends/backend1/servers/server1/url":       "http://" + whoami1IP + ":80",
		"traefik/backends/backend1/servers/server1/weight":    "10",
		"traefik/backends/backend1/servers/server2/url":       "http://" + whoami2IP + ":80",
		"traefik/backends/backend1/servers/server2/weight":    "1",
	}
	backend2 := map[string]string{
		"traefik/backends/backend2/loadbalancer/method":    "drr",
		"traefik/backends/backend2/servers/server1/url":    "http://" + whoami3IP + ":80",
		"traefik/backends/backend2/servers/server1/weight": "1",
		"traefik/backends/backend2/servers/server2/url":    "http://" + whoami4IP + ":80",
		"traefik/backends/backend2/servers/server2/weight": "2",
	}
	frontend1 := map[string]string{
		"traefik/frontends/frontend1/backend":            "backend2",
		"traefik/frontends/frontend1/entrypoints":        "http",
		"traefik/frontends/frontend1/priority":           "1",
		"traefik/frontends/frontend1/routes/test_1/rule": "Host:test.localhost",
	}
	frontend2 := map[string]string{
		"traefik/frontends/frontend2/backend":            "backend1",
		"traefik/frontends/frontend2/entrypoints":        "http",
		"traefik/frontends/frontend2/priority":           "10",
		"traefik/frontends/frontend2/routes/test_2/rule": "Path:/test",
	}
	for key, value := range backend1 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}
	for key, value := range backend2 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}
	for key, value := range frontend1 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}
	for key, value := range frontend2 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(s.kv, "traefik/frontends/frontend2/routes/test_2/rule"))
	c.Assert(err, checker.IsNil)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8081/api/providers", 60*time.Second, try.BodyContains("Path:/test"))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.localhost"

	err = try.Request(req, 500*time.Millisecond,
		try.StatusCodeIs(200),
		try.BodyContainsOr(whoami3IP, whoami4IP))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/test", nil)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 500*time.Millisecond,
		try.StatusCodeIs(200),
		try.BodyContainsOr(whoami1IP, whoami2IP))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/test2", nil)
	try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	req.Host = "test2.localhost"
	try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) TestGlobalConfiguration(c *check.C) {
	s.setupConsul(c)
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress
	err := s.kv.Put("traefik/entrypoints/http/address", []byte(":8001"), nil)
	c.Assert(err, checker.IsNil)

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(s.kv, "traefik/entrypoints/http/address"))
	c.Assert(err, checker.IsNil)

	// start traefik
	cmd, _ := s.cmdTraefik(
		withConfigFile("fixtures/simple_web.toml"),
		"--consul",
		"--consul.endpoint="+consulHost+":8500")

	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami1IP := s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress
	whoami2IP := s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress
	whoami3IP := s.composeProject.Container(c, "whoami3").NetworkSettings.IPAddress
	whoami4IP := s.composeProject.Container(c, "whoami4").NetworkSettings.IPAddress

	backend1 := map[string]string{
		"traefik/backends/backend1/circuitbreaker/expression": "NetworkErrorRatio() > 0.5",
		"traefik/backends/backend1/servers/server1/url":       "http://" + whoami1IP + ":80",
		"traefik/backends/backend1/servers/server1/weight":    "10",
		"traefik/backends/backend1/servers/server2/url":       "http://" + whoami2IP + ":80",
		"traefik/backends/backend1/servers/server2/weight":    "1",
	}
	backend2 := map[string]string{
		"traefik/backends/backend2/loadbalancer/method":    "drr",
		"traefik/backends/backend2/servers/server1/url":    "http://" + whoami3IP + ":80",
		"traefik/backends/backend2/servers/server1/weight": "1",
		"traefik/backends/backend2/servers/server2/url":    "http://" + whoami4IP + ":80",
		"traefik/backends/backend2/servers/server2/weight": "2",
	}
	frontend1 := map[string]string{
		"traefik/frontends/frontend1/backend":            "backend2",
		"traefik/frontends/frontend1/entrypoints":        "http",
		"traefik/frontends/frontend1/priority":           "1",
		"traefik/frontends/frontend1/routes/test_1/rule": "Host:test.localhost",
	}
	frontend2 := map[string]string{
		"traefik/frontends/frontend2/backend":            "backend1",
		"traefik/frontends/frontend2/entrypoints":        "http",
		"traefik/frontends/frontend2/priority":           "10",
		"traefik/frontends/frontend2/routes/test_2/rule": "Path:/test",
	}
	for key, value := range backend1 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}
	for key, value := range backend2 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}
	for key, value := range frontend1 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}
	for key, value := range frontend2 {
		err := s.kv.Put(key, []byte(value), nil)
		c.Assert(err, checker.IsNil)
	}

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(s.kv, "traefik/frontends/frontend2/routes/test_2/rule"))
	c.Assert(err, checker.IsNil)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, try.BodyContains("Path:/test"))
	c.Assert(err, checker.IsNil)

	//check
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8001/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.localhost"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) skipTestGlobalConfigurationWithClientTLS(c *check.C) {
	c.Skip("wait for relative path issue in the composefile")
	s.setupConsulTLS(c)
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress

	err := s.kv.Put("traefik/web/address", []byte(":8081"), nil)
	c.Assert(err, checker.IsNil)

	// wait for consul
	err = try.Do(60*time.Second, try.KVExists(s.kv, "traefik/web/address"))
	c.Assert(err, checker.IsNil)

	// start traefik
	cmd, _ := s.cmdTraefik(
		withConfigFile("fixtures/simple_web.toml"),
		"--consul",
		"--consul.endpoint="+consulHost+":8585",
		"--consul.tls.ca=resources/tls/ca.cert",
		"--consul.tls.cert=resources/tls/consul.cert",
		"--consul.tls.key=resources/tls/consul.key",
		"--consul.tls.insecureskipverify")

	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8081/api/providers", 60*time.Second)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulSuite) TestCommandStoreConfig(c *check.C) {
	s.setupConsul(c)
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress

	cmd, _ := s.cmdTraefik(
		"storeconfig",
		withConfigFile("fixtures/simple_web.toml"),
		"--consul.endpoint="+consulHost+":8500")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)

	// wait for traefik finish without error
	cmd.Wait()

	//CHECK
	checkmap := map[string]string{
		"/traefik/loglevel":                 "DEBUG",
		"/traefik/defaultentrypoints/0":     "http",
		"/traefik/entrypoints/http/address": ":8000",
		"/traefik/web/address":              ":8080",
		"/traefik/consul/endpoint":          consulHost + ":8500",
	}

	for key, value := range checkmap {
		var p *store.KVPair
		err = try.Do(60*time.Second, func() error {
			p, err = s.kv.Get(key)
			return err
		})
		c.Assert(err, checker.IsNil)

		c.Assert(string(p.Value), checker.Equals, value)
	}
}

type TestStruct struct {
	String string
	Int    int
}

func (s *ConsulSuite) TestDatastore(c *check.C) {
	s.setupConsul(c)
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress
	kvSource, err := staert.NewKvSource(store.CONSUL, []string{consulHost + ":8500"}, &store.Config{
		ConnectionTimeout: 10 * time.Second,
	}, "traefik")
	c.Assert(err, checker.IsNil)

	ctx := context.Background()
	datastore1, err := cluster.NewDataStore(ctx, *kvSource, &TestStruct{}, nil)
	c.Assert(err, checker.IsNil)
	datastore2, err := cluster.NewDataStore(ctx, *kvSource, &TestStruct{}, nil)
	c.Assert(err, checker.IsNil)

	setter1, _, err := datastore1.Begin()
	c.Assert(err, checker.IsNil)
	err = setter1.Commit(&TestStruct{
		String: "foo",
		Int:    1,
	})
	c.Assert(err, checker.IsNil)

	err = try.Do(3*time.Second, datastoreContains(datastore1, "foo"))
	c.Assert(err, checker.IsNil)

	err = try.Do(3*time.Second, datastoreContains(datastore2, "foo"))
	c.Assert(err, checker.IsNil)

	setter2, _, err := datastore2.Begin()
	c.Assert(err, checker.IsNil)
	err = setter2.Commit(&TestStruct{
		String: "bar",
		Int:    2,
	})
	c.Assert(err, checker.IsNil)

	err = try.Do(3*time.Second, datastoreContains(datastore1, "bar"))
	c.Assert(err, checker.IsNil)

	err = try.Do(3*time.Second, datastoreContains(datastore2, "bar"))
	c.Assert(err, checker.IsNil)

	wg := &sync.WaitGroup{}
	wg.Add(4)
	go func() {
		for i := 0; i < 100; i++ {
			setter1, _, err := datastore1.Begin()
			c.Assert(err, checker.IsNil)
			err = setter1.Commit(&TestStruct{
				String: "datastore1",
				Int:    i,
			})
			c.Assert(err, checker.IsNil)
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 100; i++ {
			setter2, _, err := datastore2.Begin()
			c.Assert(err, checker.IsNil)
			err = setter2.Commit(&TestStruct{
				String: "datastore2",
				Int:    i,
			})
			c.Assert(err, checker.IsNil)
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 100; i++ {
			test1 := datastore1.Get().(*TestStruct)
			c.Assert(test1, checker.NotNil)
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 100; i++ {
			test2 := datastore2.Get().(*TestStruct)
			c.Assert(test2, checker.NotNil)
		}
		wg.Done()
	}()
	wg.Wait()
}

func datastoreContains(datastore *cluster.Datastore, expectedValue string) func() error {
	return func() error {
		kvStruct := datastore.Get().(*TestStruct)
		if kvStruct.String != expectedValue {
			return fmt.Errorf("Got %s, wanted %s", kvStruct.String, expectedValue)
		}
		return nil
	}
}
