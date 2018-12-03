package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

type RestSuite struct{ BaseSuite }

func (s *RestSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "rest")

	s.composeProject.Start(c)
}

func (s *RestSuite) TestSimpleConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/rest/simple.toml"))

	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything.
	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	config := config.Configuration{
		Routers: map[string]*config.Router{
			"router1": {
				EntryPoints: []string{"http"},
				Middlewares: []string{},
				Service:     "service1",
				Rule:        "PathPrefix:/",
			},
		},
		Services: map[string]*config.Service{
			"service1": {
				LoadBalancer: &config.LoadBalancerService{
					Servers: []config.Server{
						{
							URL:    "http://" + s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress + ":80",
							Weight: 1,
						},
					},
				},
			},
		},
	}

	json, err := json.Marshal(config)
	c.Assert(err, checker.IsNil)

	request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8080/api/providers/rest", bytes.NewReader(json))
	c.Assert(err, checker.IsNil)

	response, err := http.DefaultClient.Do(request)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusOK)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/rest/routers", 1000*time.Millisecond, try.BodyContains("PathPrefix:/"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}
