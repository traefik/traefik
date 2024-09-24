package integration

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type RestSuite struct {
	BaseSuite
	whoamiAddr string
}

func TestRestSuite(t *testing.T) {
	suite.Run(t, new(RestSuite))
}

func (s *RestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("rest")
	s.composeUp()

	s.whoamiAddr = net.JoinHostPort(s.getComposeServiceIP("whoami1"), "80")
}

func (s *RestSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *RestSuite) TestSimpleConfigurationInsecure() {
	s.traefikCmd(withConfigFile("fixtures/rest/simple.toml"))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1000*time.Millisecond, try.BodyContains("rest@internal"))
	require.NoError(s.T(), err)

	// Expected a 404 as we did not configure anything.
	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	testCase := []struct {
		desc      string
		config    *dynamic.Configuration
		ruleMatch string
	}{
		{
			desc: "deploy http configuration",
			config: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"routerHTTP": {
							EntryPoints: []string{"web"},
							Middlewares: []string{},
							Service:     "serviceHTTP",
							Rule:        "PathPrefix(`/`)",
						},
					},
					Services: map[string]*dynamic.Service{
						"serviceHTTP": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://" + s.whoamiAddr,
									},
								},
							},
						},
					},
				},
			},
			ruleMatch: "PathPrefix(`/`)",
		},
		{
			desc: "deploy tcp configuration",
			config: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"routerTCP": {
							EntryPoints: []string{"web"},
							Service:     "serviceTCP",
							Rule:        "HostSNI(`*`)",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"serviceTCP": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: s.whoamiAddr,
									},
								},
							},
						},
					},
				},
			},
			ruleMatch: "HostSNI(`*`)",
		},
	}

	for _, test := range testCase {
		data, err := json.Marshal(test.config)
		require.NoError(s.T(), err)

		request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8080/api/providers/rest", bytes.NewReader(data))
		require.NoError(s.T(), err)

		response, err := http.DefaultClient.Do(request)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)

		err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 3*time.Second, try.BodyContains(test.ruleMatch))
		require.NoError(s.T(), err)

		err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusOK))
		require.NoError(s.T(), err)
	}
}

func (s *RestSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/rest/simple_secure.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	// Expected a 404 as we did not configure anything.
	err := try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2000*time.Millisecond, try.BodyContains("PathPrefix(`/secure`)"))
	require.NoError(s.T(), err)

	request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8080/api/providers/rest", strings.NewReader("{}"))
	require.NoError(s.T(), err)

	response, err := http.DefaultClient.Do(request)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, response.StatusCode)

	testCase := []struct {
		desc      string
		config    *dynamic.Configuration
		ruleMatch string
	}{
		{
			desc: "deploy http configuration",
			config: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"router1": {
							EntryPoints: []string{"web"},
							Middlewares: []string{},
							Service:     "service1",
							Rule:        "PathPrefix(`/`)",
						},
					},
					Services: map[string]*dynamic.Service{
						"service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://" + s.whoamiAddr,
									},
								},
							},
						},
					},
				},
			},
			ruleMatch: "PathPrefix(`/`)",
		},
		{
			desc: "deploy tcp configuration",
			config: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"router1": {
							EntryPoints: []string{"web"},
							Service:     "service1",
							Rule:        "HostSNI(`*`)",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"service1": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: s.whoamiAddr,
									},
								},
							},
						},
					},
				},
			},
			ruleMatch: "HostSNI(`*`)",
		},
	}

	for _, test := range testCase {
		data, err := json.Marshal(test.config)
		require.NoError(s.T(), err)

		request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8000/secure/api/providers/rest", bytes.NewReader(data))
		require.NoError(s.T(), err)

		response, err := http.DefaultClient.Do(request)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)

		err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains(test.ruleMatch))
		require.NoError(s.T(), err)

		err = try.GetRequest("http://127.0.0.1:8000/", time.Second, try.StatusCodeIs(http.StatusOK))
		require.NoError(s.T(), err)
	}
}
