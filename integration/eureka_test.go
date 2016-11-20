package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/containous/traefik/integration/utils"
	"github.com/go-check/check"

	checker "github.com/vdemeester/shakers"
)

// Eureka test suites (using libcompose)
type EurekaSuite struct{ BaseSuite }

func (s *EurekaSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "eureka")
	s.composeProject.Start(c)

}

func (s *EurekaSuite) TestSimpleConfiguration(c *check.C) {

	eurekaHost := s.composeProject.Container(c, "eureka").NetworkSettings.IPAddress
	whoami1Host := s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/eureka/simple.toml", struct{ EurekaHost string }{eurekaHost})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	eurekaUrl := "http://" + eurekaHost + ":8761/eureka/apps"

	// wait for eureka
	err = utils.TryRequest(eurekaUrl, 60*time.Second, func(res *http.Response) error {
		if err != nil {
			return err
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	eurekaTemplate := `
	{
    "instance": {
        "hostName": "{{ .IP }}",
        "app": "{{ .ID }}",
        "ipAddr": "{{ .IP }}",
        "status": "UP",
        "port": {
            "$": {{ .Port }},
            "@enabled": "true"
        },
        "dataCenterInfo": {
            "name": "MyOwn"
        }
    }
	}`

	tmpl, err := template.New("eurekaTemlate").Parse(eurekaTemplate)
	c.Assert(err, checker.IsNil)
	buf := new(bytes.Buffer)
	templateVars := map[string]string{
		"ID":   "tests-integration-traefik",
		"IP":   whoami1Host,
		"Port": "80",
	}
	// add in eureka
	err = tmpl.Execute(buf, templateVars)
	resp, err := http.Post(eurekaUrl+"/tests-integration-traefik", "application/json", strings.NewReader(buf.String()))
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 204)

	// wait for traefik
	err = utils.TryRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, func(res *http.Response) error {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if !strings.Contains(string(body), "Host:tests-integration-traefik") {
			return errors.New("Incorrect traefik config")
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "tests-integration-traefik"
	resp, err = client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	// TODO validate : run on 80
	resp, err = http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}
