package integration

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/containous/traefik/v2/integration/try"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

type HTTPSuite struct{ BaseSuite }

func (s *HTTPSuite) SetUpSuite(c *check.C) {
}

func (s *HTTPSuite) TestSimpleConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/http/simple.toml"))

	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything.
	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	testCase := []struct {
		desc   string
		config *dynamic.Configuration
	}{
		{
			desc: "http configuration",
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
										URL: "http://bacon:80",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCase {
		data, err := json.Marshal(test.config)
		c.Assert(err, checker.IsNil)

		ts := startTestServerWithResponse(string(data))
		defer ts.Close()

		err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 3*time.Second, try.BodyContains("bacon"))
		c.Assert(err, checker.IsNil)

		err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusOK))
		c.Assert(err, checker.IsNil)
	}
}

func startTestServerWithResponse(response string) (ts *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(response))
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
