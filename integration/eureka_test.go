package integration

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"

	checker "github.com/vdemeester/shakers"
)

// Eureka test suites (using libcompose)
type EurekaSuite struct {
	BaseSuite
	eurekaIP  string
	eurekaURL string
}

func (s *EurekaSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "eureka")
	s.composeProject.Start(c)

	eureka := s.composeProject.Container(c, "eureka")
	s.eurekaIP = eureka.NetworkSettings.IPAddress
	s.eurekaURL = "http://" + s.eurekaIP + ":8761/eureka/apps"

	// wait for eureka
	err := try.GetRequest(s.eurekaURL, 60*time.Second)
	c.Assert(err, checker.IsNil)
}

func (s *EurekaSuite) TestSimpleConfiguration(c *check.C) {

	whoami1Host := s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/eureka/simple.toml", struct{ EurekaHost string }{s.eurekaIP})
	defer os.Remove(file)
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

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

	tmpl, err := template.New("eurekaTemplate").Parse(eurekaTemplate)
	c.Assert(err, checker.IsNil)
	buf := new(bytes.Buffer)
	templateVars := map[string]string{
		"ID":   "tests-integration-traefik",
		"IP":   whoami1Host,
		"Port": "80",
	}
	// add in eureka
	err = tmpl.Execute(buf, templateVars)
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodPost, s.eurekaURL+"/tests-integration-traefik", strings.NewReader(buf.String()))
	c.Assert(err, checker.IsNil)
	req.Header.Set("Content-Type", "application/json")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusNoContent))
	c.Assert(err, checker.IsNil)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, try.BodyContains("Host:tests-integration-traefik"))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "tests-integration-traefik"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}
