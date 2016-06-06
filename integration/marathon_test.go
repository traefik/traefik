package main

import (
	"net/http"
	"os/exec"
	"time"

	"github.com/go-check/check"

	checker "github.com/vdemeester/shakers"
)

// Marathon test suites (using libcompose)
type MarathonSuite struct{ BaseSuite }

func (s *MarathonSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "marathon")
	s.composeProject.Start(c)
	// wait for marathon
	// err := utils.TryRequest("http://127.0.0.1:8080/ping", 60*time.Second, func(res *http.Response) error {
	// 	body, err := ioutil.ReadAll(res.Body)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if !strings.Contains(string(body), "ping") {
	// 		return errors.New("Incorrect marathon config")
	// 	}
	// 	return nil
	// })
	// c.Assert(err, checker.IsNil)
}

func (s *MarathonSuite) TestSimpleConfiguration(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/marathon/simple.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)
	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}
