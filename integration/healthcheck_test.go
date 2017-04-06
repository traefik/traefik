package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/containous/traefik/integration/utils"
	"github.com/go-check/check"

	checker "github.com/vdemeester/shakers"
)

// HealchCheck test suites (using libcompose)
type HealchCheckSuite struct{ BaseSuite }

func (s *HealchCheckSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "healthcheck")
	s.composeProject.Start(c)

}

func (s *HealchCheckSuite) TestSimpleConfiguration(c *check.C) {

	whoami1Host := s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress
	whoami2Host := s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/healthcheck/simple.toml", struct {
		Server1 string
		Server2 string
	}{whoami1Host, whoami2Host})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = utils.TryRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, func(res *http.Response) error {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if !strings.Contains(string(body), "Host:test.localhost") {
			return errors.New("Incorrect traefik config: " + string(body))
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/health", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.localhost"

	resp, err := client.Do(req)
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	resp, err = client.Do(req)
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	healthReq, err := http.NewRequest("POST", "http://"+whoami1Host+"/health", bytes.NewBuffer([]byte("500")))
	c.Assert(err, checker.IsNil)
	_, err = client.Do(healthReq)
	c.Assert(err, checker.IsNil)

	time.Sleep(time.Second * 3)

	resp, err = client.Do(req)
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	resp, err = client.Do(req)
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	// TODO validate : run on 80
	resp, err = http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}
