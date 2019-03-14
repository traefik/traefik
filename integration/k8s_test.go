package integration

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// K8sSuite
type K8sSuite struct{ BaseSuite }

func (s *K8sSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "k8s")
	s.composeProject.Start(c)

	abs, err := filepath.Abs("./fixtures/k8s/kubeconfig.yaml")
	c.Assert(err, checker.IsNil)

	err = try.Do(60*time.Second, try.DoCondition(func() error {
		_, err := os.Stat(abs)
		return err
	}))
	c.Assert(err, checker.IsNil)

	err = os.Setenv("KUBECONFIG", abs)
	c.Assert(err, checker.IsNil)
}

func (s *K8sSuite) TearDownSuite(c *check.C) {
	s.composeProject.Stop(c)

	err := os.Remove("./fixtures/k8s/kubeconfig.yaml")
	if err != nil {
		c.Log(err)
	}
	err = os.Remove("./fixtures/k8s/coredns.yaml")
	if err != nil {
		c.Log(err)
	}
	err = os.Remove("./fixtures/k8s/traefik.yaml")
	if err != nil {
		c.Log(err)
	}
}

func (s *K8sSuite) TestIngressSimple(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/k8s_default.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("Host(`whoami.test`)"))
	c.Assert(err, checker.IsNil)
}

func (s *K8sSuite) TestCRDSimple(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/k8s_crd.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("Host(`foo.com`)"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("PathPrefix(`/tobestripped`)"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/kubernetescrd/routers", 1*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("default/stripprefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/kubernetescrd/middlewares", 1*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("stripprefix"))
	c.Assert(err, checker.IsNil)
}
