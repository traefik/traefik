package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-check/check"

	"errors"
	"github.com/containous/traefik/integration/utils"
	checker "github.com/vdemeester/shakers"
)

// ACME test suites (using libcompose)
type AcmeSuite struct {
	BaseSuite
}

func (s *AcmeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "boulder")
	s.composeProject.Start(c)

	boulderHost := s.composeProject.Container(c, "boulder").NetworkSettings.IPAddress

	// wait for boulder
	err := utils.Try(120*time.Second, func() error {
		resp, err := http.Get("http://" + boulderHost + ":4000/directory")
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New("Expected http 200 from boulder")
		}
		return nil
	})

	c.Assert(err, checker.IsNil)
}

func (s *AcmeSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *AcmeSuite) TestRetrieveAcmeCertificate(c *check.C) {
	boulderHost := s.composeProject.Container(c, "boulder").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/acme/acme.toml", struct{ BoulderHost string }{boulderHost})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	backend := startTestServer("9010", 200)
	defer backend.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// wait for traefik (generating acme account take some seconds)
	err = utils.Try(30*time.Second, func() error {
		_, err := client.Get("https://127.0.0.1:5001")
		if err != nil {
			return err
		}
		return nil
	})
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
	c.Assert(resp.StatusCode, checker.Equals, 200)
}
