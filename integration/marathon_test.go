package main

import (
	"net/http"
	"os/exec"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// Marathon test suites (using libcompose)
type MarathonSuite struct{ BaseSuite }

func (s *MarathonSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "marathon")
	s.composeProject.Start(c)

	// FIXME Doesn't work...
	//// "github.com/gambol99/go-marathon"
	//config := marathon.NewDefaultConfig()
	//
	//marathonClient, err := marathon.NewClient(config)
	//if err != nil {
	//	c.Fatalf("Error creating Marathon client. %v", err)
	//}
	//
	//// Wait for Marathon to elect itself leader
	//err = try.Do(30*time.Second, func() error {
	//	leader, err := marathonClient.Leader()
	//
	//	if err != nil || len(leader) == 0 {
	//		return fmt.Errorf("Leader not found. %v", err)
	//	}
	//
	//	return nil
	//})
	//
	//c.Assert(err, checker.IsNil)
}

func (s *MarathonSuite) TestSimpleConfiguration(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/marathon/simple.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))

	c.Assert(err, checker.IsNil)
}
