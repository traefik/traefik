package integration

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// SimpleSuite
type SimpleSuite struct{ BaseSuite }

func (s *SimpleSuite) TestInvalidConfigShouldFail(c *check.C) {
	cmd, output := s.cmdTraefik(withConfigFile("fixtures/invalid_configuration.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.Do(500*time.Millisecond, func() error {
		expected := "Near line 0 (last key parsed ''): bare keys cannot contain '{'"
		actual := output.String()

		if !strings.Contains(actual, expected) {
			return fmt.Errorf("got %s, wanted %s", actual, expected)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestSimpleDefaultConfig(c *check.C) {
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/simple_default.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWithWebConfig(c *check.C) {
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/simple_web.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestDefaultEntryPoints(c *check.C) {
	cmd, output := s.cmdTraefik("--debug")

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.Do(500*time.Millisecond, func() error {
		expected := `\"DefaultEntryPoints\":[\"http\"]`
		actual := output.String()

		if !strings.Contains(actual, expected) {
			return fmt.Errorf("got %s, wanted %s", actual, expected)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestPrintHelp(c *check.C) {
	cmd, output := s.cmdTraefik("--help")

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.Do(500*time.Millisecond, func() error {
		expected := "Usage:"
		notExpected := "panic:"
		actual := output.String()

		if strings.Contains(actual, notExpected) {
			return fmt.Errorf("got %s", actual)
		}
		if !strings.Contains(actual, expected) {
			return fmt.Errorf("got %s, wanted %s", actual, expected)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestRequestAcceptGraceTimeout(c *check.C) {
	s.createComposeProject(c, "reqacceptgrace")
	s.composeProject.Start(c)

	whoami := "http://" + s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress + ":80"

	file := s.adaptFile(c, "fixtures/reqacceptgrace.toml", struct {
		Server string
	}{whoami})
	defer os.Remove(file)
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	// Make sure exposed service is ready.
	err = try.GetRequest("http://127.0.0.1:8000/service", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Check that /ping endpoint is responding with 200.
	err = try.GetRequest("http://127.0.0.1:8001/ping", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Send SIGTERM to Traefik.
	proc, err := os.FindProcess(cmd.Process.Pid)
	c.Assert(err, checker.IsNil)
	err = proc.Signal(syscall.SIGTERM)
	c.Assert(err, checker.IsNil)

	// Give Traefik time to process the SIGTERM and send a request half-way
	// into the request accepting grace period, by which requests should
	// still get served.
	time.Sleep(5 * time.Second)
	resp, err := http.Get("http://127.0.0.1:8000/service")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, checker.Equals, http.StatusOK)

	// ping endpoint should now return a Service Unavailable.
	resp, err = http.Get("http://127.0.0.1:8001/ping")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, checker.Equals, http.StatusServiceUnavailable)

	// Expect Traefik to shut down gracefully once the request accepting grace
	// period has elapsed.
	waitErr := make(chan error)
	go func() {
		waitErr <- cmd.Wait()
	}()

	select {
	case err := <-waitErr:
		c.Assert(err, checker.IsNil)
	case <-time.After(10 * time.Second):
		// By now we are ~5 seconds out of the request accepting grace period
		// (start + 5 seconds sleep prior to the mid-grace period request +
		// 10 seconds timeout = 15 seconds > 10 seconds grace period).
		// Something must have gone wrong if we still haven't terminated at
		// this point.
		c.Fatal("Traefik did not terminate in time")
	}
}

func (s *SimpleSuite) TestApiOnSameEntryPoint(c *check.C) {
	s.createComposeProject(c, "base")
	s.composeProject.Start(c)

	cmd, output := s.traefikCmd("--defaultEntryPoints=http", "--entryPoints=Name:http Address::8000", "--api.entryPoint=http", "--debug", "--docker")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/test", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestStatsWithMultipleEntryPoint(c *check.C) {
	s.createComposeProject(c, "stats")
	s.composeProject.Start(c)

	whoami1 := "http://" + s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress + ":80"
	whoami2 := "http://" + s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress + ":80"

	file := s.adaptFile(c, "fixtures/simple_stats.toml", struct {
		Server1 string
		Server2 string
	}{whoami1, whoami2})
	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/health", 1*time.Second, try.BodyContains(`"total_status_code_count":{"200":2}`))
	c.Assert(err, checker.IsNil)

}

func (s *SimpleSuite) TestNoAuthOnPing(c *check.C) {
	s.createComposeProject(c, "base")
	s.composeProject.Start(c)

	cmd, output := s.traefikCmd(withConfigFile("./fixtures/simple_auth.toml"))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8001/api", 1*time.Second, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8001/ping", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWebCompatibilityWithoutPath(c *check.C) {

	s.createComposeProject(c, "base")
	s.composeProject.Start(c)

	cmd, output := s.traefikCmd("--defaultEntryPoints=http", "--entryPoints=Name:http Address::8000", "--web", "--debug", "--docker")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/test", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWebCompatibilityWithPath(c *check.C) {

	s.createComposeProject(c, "base")
	s.composeProject.Start(c)

	cmd, output := s.traefikCmd("--defaultEntryPoints=http", "--entryPoints=Name:http Address::8000", "--web.path=/test", "--debug", "--docker")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/notfound", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/test/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/test/ping", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/test/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestDefaultEntrypointHTTP(c *check.C) {

	s.createComposeProject(c, "base")
	s.composeProject.Start(c)

	cmd, output := s.traefikCmd("--entryPoints=Name:http Address::8000", "--debug", "--docker", "--api")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWithUnexistingEntrypoint(c *check.C) {

	s.createComposeProject(c, "base")
	s.composeProject.Start(c)

	cmd, output := s.traefikCmd("--defaultEntryPoints=https,http", "--entryPoints=Name:http Address::8000", "--debug", "--docker", "--api")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestMetricsPrometheusDefaultEntrypoint(c *check.C) {

	s.createComposeProject(c, "base")
	s.composeProject.Start(c)

	cmd, output := s.traefikCmd("--defaultEntryPoints=http", "--entryPoints=Name:http Address::8000", "--web", "--web.metrics.prometheus.buckets=0.1,0.3,1.2,5.0", "--docker", "--debug")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestMultipleProviderSameBackendName(c *check.C) {

	s.createComposeProject(c, "base")
	s.composeProject.Start(c)
	ipWhoami01 := s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress
	ipWhoami02 := s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/multiple_provider.toml", struct{ IP string }{
		IP: ipWhoami02,
	})
	defer os.Remove(file)
	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.BodyContains(ipWhoami01))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/file", 1*time.Second, try.BodyContains(ipWhoami02))
	c.Assert(err, checker.IsNil)

}

func (s *SimpleSuite) TestDontKeepTrailingSlash(c *check.C) {
	file := s.adaptFile(c, "fixtures/keep_trailing_slash.toml", struct {
		KeepTrailingSlash bool
	}{false})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	oldCheckRedirect := http.DefaultClient.CheckRedirect
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	err = try.GetRequest("http://127.0.0.1:8000/test/foo/", 1*time.Second, try.StatusCodeIs(http.StatusMovedPermanently))
	c.Assert(err, checker.IsNil)

	http.DefaultClient.CheckRedirect = oldCheckRedirect
}

func (s *SimpleSuite) TestKeepTrailingSlash(c *check.C) {
	file := s.adaptFile(c, "fixtures/keep_trailing_slash.toml", struct {
		KeepTrailingSlash bool
	}{true})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	oldCheckRedirect := http.DefaultClient.CheckRedirect
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	err = try.GetRequest("http://127.0.0.1:8000/test/foo/", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	http.DefaultClient.CheckRedirect = oldCheckRedirect
}
