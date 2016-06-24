package main

import (
	"net/http"
	"os/exec"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/go-check/check"

	"errors"
	"github.com/containous/traefik/integration/utils"
	checker "github.com/vdemeester/shakers"
	"io/ioutil"
	"os"
	"strings"
)

// Consul test suites (using libcompose)
type ConsulSuite struct {
	BaseSuite
	kv store.Store
}

func (s *ConsulSuite) SetUpTest(c *check.C) {
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
	err = utils.Try(60*time.Second, func() error {
		_, err := kv.Exists("test")
		if err != nil {
			return err
		}
		return nil
	})
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
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/consul/simple.toml", struct{ ConsulHost string }{consulHost})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}

func (s *ConsulSuite) TestNominalConfiguration(c *check.C) {
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/consul/simple.toml", struct{ ConsulHost string }{consulHost})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami1 := s.composeProject.Container(c, "whoami1")
	whoami2 := s.composeProject.Container(c, "whoami2")
	whoami3 := s.composeProject.Container(c, "whoami3")
	whoami4 := s.composeProject.Container(c, "whoami4")

	backend1 := map[string]string{
		"traefik/backends/backend1/circuitbreaker/expression": "NetworkErrorRatio() > 0.5",
		"traefik/backends/backend1/servers/server1/url":       "http://" + whoami1.NetworkSettings.IPAddress + ":80",
		"traefik/backends/backend1/servers/server1/weight":    "10",
		"traefik/backends/backend1/servers/server2/url":       "http://" + whoami2.NetworkSettings.IPAddress + ":80",
		"traefik/backends/backend1/servers/server2/weight":    "1",
	}
	backend2 := map[string]string{
		"traefik/backends/backend2/loadbalancer/method":    "drr",
		"traefik/backends/backend2/servers/server1/url":    "http://" + whoami3.NetworkSettings.IPAddress + ":80",
		"traefik/backends/backend2/servers/server1/weight": "1",
		"traefik/backends/backend2/servers/server2/url":    "http://" + whoami4.NetworkSettings.IPAddress + ":80",
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
	err = utils.Try(60*time.Second, func() error {
		_, err := s.kv.Exists("traefik/frontends/frontend2/routes/test_2/rule")
		if err != nil {
			return err
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	// wait for traefik
	err = utils.TryRequest("http://127.0.0.1:8081/api/providers", 60*time.Second, func(res *http.Response) error {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if !strings.Contains(string(body), "Path:/test") {
			return errors.New("Incorrect traefik config")
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.localhost"
	response, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, 200)

	body, err := ioutil.ReadAll(response.Body)
	c.Assert(err, checker.IsNil)
	if !strings.Contains(string(body), whoami3.NetworkSettings.IPAddress) &&
		!strings.Contains(string(body), whoami4.NetworkSettings.IPAddress) {
		c.Fail()
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:8000/test", nil)
	c.Assert(err, checker.IsNil)
	response, err = client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, 200)

	body, err = ioutil.ReadAll(response.Body)
	c.Assert(err, checker.IsNil)
	if !strings.Contains(string(body), whoami1.NetworkSettings.IPAddress) &&
		!strings.Contains(string(body), whoami2.NetworkSettings.IPAddress) {
		c.Fail()
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:8000/test2", nil)
	resp, err := client.Do(req)
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)

	req, err = http.NewRequest("GET", "http://127.0.0.1:8000/", nil)
	req.Host = "test2.localhost"
	resp, err = client.Do(req)
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}

func (s *ConsulSuite) TestGlobalConfiguration(c *check.C) {
	consulHost := s.composeProject.Container(c, "consul").NetworkSettings.IPAddress
	err := s.kv.Put("traefik/entrypoints/http/address", []byte(":8001"), nil)
	c.Assert(err, checker.IsNil)

	// wait for consul
	err = utils.Try(60*time.Second, func() error {
		_, err := s.kv.Exists("traefik/entrypoints/http/address")
		if err != nil {
			return err
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	// start traefik
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/simple_web.toml", "--consul", "--consul.endpoint="+consulHost+":8500")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami1 := s.composeProject.Container(c, "whoami1")
	whoami2 := s.composeProject.Container(c, "whoami2")
	whoami3 := s.composeProject.Container(c, "whoami3")
	whoami4 := s.composeProject.Container(c, "whoami4")

	backend1 := map[string]string{
		"traefik/backends/backend1/circuitbreaker/expression": "NetworkErrorRatio() > 0.5",
		"traefik/backends/backend1/servers/server1/url":       "http://" + whoami1.NetworkSettings.IPAddress + ":80",
		"traefik/backends/backend1/servers/server1/weight":    "10",
		"traefik/backends/backend1/servers/server2/url":       "http://" + whoami2.NetworkSettings.IPAddress + ":80",
		"traefik/backends/backend1/servers/server2/weight":    "1",
	}
	backend2 := map[string]string{
		"traefik/backends/backend2/loadbalancer/method":    "drr",
		"traefik/backends/backend2/servers/server1/url":    "http://" + whoami3.NetworkSettings.IPAddress + ":80",
		"traefik/backends/backend2/servers/server1/weight": "1",
		"traefik/backends/backend2/servers/server2/url":    "http://" + whoami4.NetworkSettings.IPAddress + ":80",
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
	err = utils.Try(60*time.Second, func() error {
		_, err := s.kv.Exists("traefik/frontends/frontend2/routes/test_2/rule")
		if err != nil {
			return err
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	// wait for traefik
	err = utils.TryRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, func(res *http.Response) error {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if !strings.Contains(string(body), "Path:/test") {
			return errors.New("Incorrect traefik config")
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	//check
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8001/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.localhost"
	response, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, 200)
}
