package main

import (
	"net/http"
	"os/exec"
	"time"

	"github.com/go-check/check"

	checker "github.com/vdemeester/shakers"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
	"io/ioutil"
	"strings"
)

// Etcd test suites (using libcompose)
type EtcdSuite struct {
	BaseSuite
	kv store.Store
}

func (s *EtcdSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "etcd")
	s.composeProject.Start(c)

	time.Sleep(3000 * time.Millisecond)

	etcd.Register()
	kv, err := libkv.NewStore(
		store.ETCD,
		[]string{"localhost:4001"},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		c.Fatal("Cannot create store etcd")
	}

	resp, err := http.Get("http://127.0.0.1:4001/v2/keys")

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)
	s.kv = kv
}

func (s *EtcdSuite) TestSimpleConfiguration(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/etcd/simple.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(1000 * time.Millisecond)
	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}

func (s *EtcdSuite) TestNominalConfiguration(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/etcd/simple.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(1000 * time.Millisecond)
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)

	whoami1 := s.composeProject.Container(c, "whoami1")
	whoami2 := s.composeProject.Container(c, "whoami2")
	whoami3 := s.composeProject.Container(c, "whoami3")
	whoami4 := s.composeProject.Container(c, "whoami4")

	backend1 := map[string]string{
		"/traefik/backends/backend1/circuitbreaker/expression": "NetworkErrorRatio() > 0.5",
		"/traefik/backends/backend1/servers/server1/url":       "http://" + whoami1.NetworkSettings.IPAddress + ":80",
		"/traefik/backends/backend1/servers/server1/weight":    "10",
		"/traefik/backends/backend1/servers/server2/url":       "http://" + whoami2.NetworkSettings.IPAddress + ":80",
		"/traefik/backends/backend1/servers/server2/weight":    "1",
	}
	backend2 := map[string]string{
		"/traefik/backends/backend2/loadbalancer/method":    "drr",
		"/traefik/backends/backend2/servers/server1/url":    "http://" + whoami3.NetworkSettings.IPAddress + ":80",
		"/traefik/backends/backend2/servers/server1/weight": "1",
		"/traefik/backends/backend2/servers/server2/url":    "http://" + whoami4.NetworkSettings.IPAddress + ":80",
		"/traefik/backends/backend2/servers/server2/weight": "2",
	}
	frontend1 := map[string]string{
		"/traefik/frontends/frontend1/backend":            "backend2",
		"/traefik/frontends/frontend1/entrypoints":        "http",
		"/traefik/frontends/frontend1/routes/test_1/rule": "Host:test.localhost",
	}
	frontend2 := map[string]string{
		"/traefik/frontends/frontend2/backend":            "backend1",
		"/traefik/frontends/frontend2/entrypoints":        "http",
		"/traefik/frontends/frontend2/routes/test_2/rule": "Path:/test",
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

	time.Sleep(3000 * time.Millisecond)

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
	req.Host = "test2.localhost"
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)

	req, err = http.NewRequest("GET", "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}
