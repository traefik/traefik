package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type ThrottlingSuite struct{ BaseSuite }

func TestThrottlingSuite(t *testing.T) {
	suite.Run(t, new(ThrottlingSuite))
}

func (s *ThrottlingSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.createComposeProject("rest")
	s.composeUp()
}

func (s *ThrottlingSuite) TestThrottleConfReload() {
	s.traefikCmd(withConfigFile("fixtures/throttling/simple.toml"))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.BodyContains("rest@internal"))
	require.NoError(s.T(), err)

	// Expected a 404 as we did not configure anything.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	config := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{},
			Services: map[string]*dynamic.Service{
				"serviceHTTP": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{
							{
								URL: "http://" + s.getComposeServiceIP("whoami1") + ":80",
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

	for i := range confChanges {
		config.HTTP.Routers[fmt.Sprintf("routerHTTP%d", i)] = router
		data, err := json.Marshal(config)
		require.NoError(s.T(), err)

		request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8080/api/providers/rest", bytes.NewReader(data))
		require.NoError(s.T(), err)

		response, err := http.DefaultClient.Do(request)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)
		time.Sleep(200 * time.Millisecond)
	}

	reloadsRegexp := regexp.MustCompile(`traefik_config_reloads_total (\d*)\n`)

	resp, err := http.Get("http://127.0.0.1:8080/metrics")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)

	fields := reloadsRegexp.FindStringSubmatch(string(body))
	assert.Len(s.T(), fields, 2)

	reloads, err := strconv.Atoi(fields[1])
	if err != nil {
		panic(err)
	}

	// The test tries to trigger a config reload with the REST API every 200ms,
	// 10 times (so for 2s in total).
	// Therefore the throttling (set at 400ms for this test) should only let
	// (2s / 400 ms =) 5 config reloads happen in theory.
	// In addition, we have to take into account the extra config reload from the internal provider (5 + 1).
	assert.LessOrEqual(s.T(), reloads, 6)
}
