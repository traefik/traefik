package integration

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	checker "github.com/vdemeester/shakers"
)

type HTTPSuite struct{ BaseSuite }

func (s *HTTPSuite) TestSimpleConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/http/simple.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)

	defer cmd.Process.Kill()

	// Expect a 404 as we configured nothing.
	err = try.GetRequest("http://127.0.0.1:8000/", time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

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
						PassHostHeader: boolRef(true),
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
	c.Assert(err, checker.IsNil)

	server := startTestServerWithResponse(configData)
	defer server.Close()

	// Expect configuration to be applied.
	err = try.GetRequest("http://127.0.0.1:9090/api/rawdata", 3*time.Second, try.BodyContains("routerHTTP@http", "serviceHTTP@http", "http://bacon:80"))
	c.Assert(err, checker.IsNil)
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

func boolRef(b bool) *bool {
	return &b
}
