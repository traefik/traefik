package integration

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type HTTPSuite struct{ BaseSuite }

func TestHTTPSuite(t *testing.T) {
	suite.Run(t, new(HTTPSuite))
}

func (s *HTTPSuite) TestSimpleConfiguration() {
	s.traefikCmd(withConfigFile("fixtures/http/simple.toml"))

	// Expect a 404 as we configured nothing.
	err := try.GetRequest("http://127.0.0.1:8000/", time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	// Provide a configuration, fetched by Traefik provider.
	configuration := &dynamic.Configuration{
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
						PassHostHeader: pointer(true),
						Servers: []dynamic.Server{
							{
								URL: "http://bacon:80",
							},
						},
					},
				},
			},
		},
	}

	configData, err := json.Marshal(configuration)
	require.NoError(s.T(), err)

	server := startTestServerWithResponse(configData)
	defer server.Close()

	// Expect configuration to be applied.
	err = try.GetRequest("http://127.0.0.1:9090/api/rawdata", 3*time.Second, try.BodyContains("routerHTTP@http", "serviceHTTP@http", "http://bacon:80"))
	require.NoError(s.T(), err)
}

func startTestServerWithResponse(response []byte) (ts *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(response)
	})
	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}

	ts = &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	ts.Start()
	return ts
}

func pointer[T any](v T) *T { return &v }
