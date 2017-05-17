package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// ACME test suites (using libcompose)
type AcmeSuite struct {
	BaseSuite
	boulderIP string
}

func (s *AcmeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "boulder")
	s.composeProject.Start(c)

	s.boulderIP = s.composeProject.Container(c, "boulder").NetworkSettings.IPAddress

	// wait for boulder
	err := try.GetRequest("http://"+s.boulderIP+":4000/directory", 120*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *AcmeSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *AcmeSuite) TestRetrieveAcmeCertificate(c *check.C) {
	file := s.adaptFile(c, "fixtures/acme/acme.toml", struct{ BoulderHost string }{s.boulderIP})
	defer os.Remove(file)
	cmd, output := s.cmdTraefikWithConfigFile(file)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	backend := startTestServer("9010", http.StatusOK)
	defer backend.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// wait for traefik (generating acme account take some seconds)
	err = try.Do(90*time.Second, func() error {
		_, err := client.Get("https://127.0.0.1:5001")
		return err
	})
	// TODO: waiting a refactor of integration tests
	s.displayTraefikLog(c, output)
	c.Assert(err, checker.IsNil)

	tr = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "traefik.acme.wtf",
		},
	}
	client = &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", "https://127.0.0.1:5001/", nil)
	req.Host = "traefik.acme.wtf"
	req.Header.Set("Host", "traefik.acme.wtf")
	req.Header.Set("Accept", "*/*")
	resp, err := client.Do(req)
	c.Assert(err, checker.IsNil)
	// Expected a 200
	c.Assert(resp.StatusCode, checker.Equals, http.StatusOK)
}
