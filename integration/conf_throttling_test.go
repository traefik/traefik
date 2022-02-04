package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	checker "github.com/vdemeester/shakers"
)

type ThrottlingSuite struct{ BaseSuite }

func (s *ThrottlingSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "rest")
	s.composeUp(c)
}

func (s *ThrottlingSuite) TestThrottleConfReload(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/throttling/simple.toml"))

	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1000*time.Millisecond, try.BodyContains("rest@internal"))
	c.Assert(err, checker.IsNil)

	// Expected a 404 as we did not configure anything.
	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	config := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{},
			Services: map[string]*dynamic.Service{
				"serviceHTTP": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{
							{
								URL: "http://" + s.getComposeServiceIP(c, "whoami1") + ":80",
							},
						},
					},
				},
			},
		},
	}

	router := &dynamic.Router{
		EntryPoints: []string{"web"},
		Middlewares: []string{},
		Service:     "serviceHTTP",
		Rule:        "PathPrefix(`/`)",
	}

	confChanges := 10

	for i := 0; i < confChanges; i++ {
		config.HTTP.Routers[fmt.Sprintf("routerHTTP%d", i)] = router
		data, err := json.Marshal(config)
		c.Assert(err, checker.IsNil)

		request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8080/api/providers/rest", bytes.NewReader(data))
		c.Assert(err, checker.IsNil)

		response, err := http.DefaultClient.Do(request)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
		time.Sleep(200 * time.Millisecond)
	}

	reloadsRegexp := regexp.MustCompile(`traefik_config_reloads_total (\d*)\n`)

	resp, err := http.Get("http://127.0.0.1:8080/metrics")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	fields := reloadsRegexp.FindStringSubmatch(string(body))
	c.Assert(len(fields), checker.Equals, 2)

	reloads, err := strconv.Atoi(fields[1])
	if err != nil {
		panic(err)
	}

	// The test tries to trigger a config reload with the REST API every 200ms,
	// 10 times (so for 2s in total).
	// Therefore the throttling (set at 400ms for this test) should only let
	// (2s / 400 ms =) 5 config reloads happen in theory.
	// In addition, we have to take into account the extra config reload from the internal provider (5 + 1).
	c.Assert(reloads, checker.LessOrEqualThan, 6)
}
