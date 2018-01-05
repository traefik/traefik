package integration

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/provider/label"
	"github.com/go-check/check"
	d "github.com/libkermit/docker"
	"github.com/libkermit/docker-check"
	"github.com/mattn/go-shellwords"
	checker "github.com/vdemeester/shakers"
)

const (
	traefikTestLogFile       = "traefik.log"
	traefikTestAccessLogFile = "access.log"
)

var requireImages = map[string]string{
	"emilevauge/whoami": "latest",
}

// AccessLogSuite
type AccessLogSuite struct {
	BaseSuite
	project *docker.Project
}

type accessLogValue struct {
	formatOnly  bool
	code        string
	value       string
	backendName string
}

func (s *AccessLogSuite) SetUpSuite(c *check.C) {
	s.project = docker.NewProjectFromEnv(c)

	// Pull required images
	for repository, tag := range requireImages {
		image := fmt.Sprintf("%s:%s", repository, tag)
		s.project.Pull(c, image)
	}
}

func (s *AccessLogSuite) TearDownTest(c *check.C) {
	s.project.Clean(c, os.Getenv("CIRCLECI") != "")
}

func (s *AccessLogSuite) TestAccessLog(c *check.C) {
	// Ensure working directory is clean
	ensureWorkingDirectoryIsClean()

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	defer os.Remove(traefikTestAccessLogFile)
	defer os.Remove(traefikTestLogFile)

	checkStatsForLogFile(c)

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Start test servers
	ts1 := startAccessLogServer(8081)
	defer ts1.Close()
	ts2 := startAccessLogServer(8082)
	defer ts2.Close()
	ts3 := startAccessLogServer(8083)
	defer ts3.Close()

	// Make some requests
	err = try.GetRequest("http://127.0.0.1:8000/test1", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8000/test2", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8000/test2", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogOutput(err, c)

	c.Assert(count, checker.GreaterOrEqualThan, 3)

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestInternalAccessLog(c *check.C) {
	// Ensure working directory is clean
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "401",
			value:       "Auth for frontend-Host-frontend-auth-docker-local",
			backendName: "-",
		},
		{
			formatOnly:  false,
			code:        "401",
			value:       "Auth for entrypoint",
			backendName: "-",
		},
		{
			formatOnly:  false,
			code:        "302",
			value:       "entrypoint redirect for frontend-Host-whoami-docker-local",
			backendName: "-",
		},
		{
			formatOnly: true,
		},
		{
			formatOnly:  false,
			code:        "302",
			value:       "frontend redirect for frontend-Path-",
			backendName: "-",
		},
		{
			formatOnly: true,
		},
		{
			formatOnly: true,
		},
		{
			formatOnly: true,
		},
		{
			formatOnly:  false,
			code:        "429",
			value:       "ratelimit",
			backendName: "-",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	defer os.Remove(traefikTestAccessLogFile)
	defer os.Remove(traefikTestLogFile)

	checkStatsForLogFile(c)

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Start test servers
	ts1 := startAccessLogServer(8081)
	defer ts1.Close()
	ts2 := startAccessLogServer(8082)
	defer ts2.Close()
	ts3 := startAccessLogServer(8083)
	defer ts3.Close()

	s.testAuthFrontEnd(c)
	s.testAuthEntrypoint(c)
	s.testEntrypointRedirect(c)
	s.testFrontendRedirect(c)
	s.testRateLimit(c)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) testAuthFrontEnd(c *check.C) {
	labels := map[string]string{
		label.TraefikEnable:              "true",
		label.TraefikPort:                "80",
		label.TraefikBackend:             "backend3",
		label.TraefikFrontendEntryPoints: "httpFrontendAuth",
		label.TraefikFrontendRule:        "Host:frontend.auth.docker.local",
		label.TraefikFrontendAuthBasic:   "test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/",
	}
	s.startContainerWithNameAndLabels(c, "whoami0", "emilevauge/whoami", labels, "")

	// Test auth frontend
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8006/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend.auth.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	c.Assert(err, checker.IsNil)

	s.project.Stop(c, "whoami0")
	s.project.Remove(c, "whoami0")
}

func (s *AccessLogSuite) testAuthEntrypoint(c *check.C) {
	labels := map[string]string{
		label.TraefikEnable:              "true",
		label.TraefikPort:                "80",
		label.TraefikBackend:             "backend3",
		label.TraefikFrontendEntryPoints: "httpAuth",
		label.TraefikFrontendRule:        "Host:entrypoint.auth.docker.local",
	}
	s.startContainerWithNameAndLabels(c, "whoami0", "emilevauge/whoami", labels, "")

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8004/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.auth.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	c.Assert(err, checker.IsNil)

	s.project.Stop(c, "whoami0")
	s.project.Remove(c, "whoami0")
}

func (s *AccessLogSuite) testEntrypointRedirect(c *check.C) {
	labels := map[string]string{
		label.TraefikEnable:              "true",
		label.TraefikPort:                "80",
		label.TraefikBackend:             "backend3",
		label.TraefikFrontendEntryPoints: "httpRedirect",
		label.TraefikFrontendRule:        "Path:/test1",
	}
	s.startContainerWithNameAndLabels(c, "whoami0", "emilevauge/whoami", labels, "")

	// Test entrypoint redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8001/test1", nil)
	c.Assert(err, checker.IsNil)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	s.project.Stop(c, "whoami0")
	s.project.Remove(c, "whoami0")
}

func (s *AccessLogSuite) testFrontendRedirect(c *check.C) {
	labels := map[string]string{
		label.TraefikEnable:                     "true",
		label.TraefikPort:                       "80",
		label.TraefikBackend:                    "backend3",
		label.TraefikFrontendEntryPoints:        "frontendRedirect,http",
		label.TraefikFrontendRule:               "Path:/test1",
		label.TraefikFrontendRedirectEntryPoint: "http",
	}
	s.startContainerWithNameAndLabels(c, "whoami0", "emilevauge/whoami", labels, "")

	// Test frontend redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8005/test1", nil)
	c.Assert(err, checker.IsNil)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	s.project.Stop(c, "whoami0")
	s.project.Remove(c, "whoami0")
}

func (s *AccessLogSuite) testRateLimit(c *check.C) {
	rateLimitLabel := label.Prefix + label.BaseFrontendRateLimit + "powpow."
	labels := map[string]string{
		label.TraefikEnable:                           "true",
		label.TraefikPort:                             "80",
		label.TraefikBackend:                          "backend3",
		label.TraefikFrontendEntryPoints:              "http",
		label.TraefikFrontendRule:                     "Host:ratelimit.docker.local",
		label.TraefikFrontendRateLimitExtractorFunc:   "client.ip",
		rateLimitLabel + label.SuffixRateLimitPeriod:  "3s",
		rateLimitLabel + label.SuffixRateLimitAverage: "1",
		rateLimitLabel + label.SuffixRateLimitBurst:   "2",
	}

	s.startContainerWithNameAndLabels(c, "whoami0", "emilevauge/whoami", labels, "")

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "ratelimit.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests), try.HasBody())
	c.Assert(err, checker.IsNil)

	s.project.Stop(c, "whoami0")
	s.project.Remove(c, "whoami0")
}

func checkNoOtherTraefikProblems(traefikLog []byte, err error, c *check.C) {
	traefikLog, err = ioutil.ReadFile(traefikTestLogFile)
	c.Assert(err, checker.IsNil)
	if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
		c.Assert(traefikLog, checker.HasLen, 0)
	}
}

func checkAccessLogOutput(err error, c *check.C) int {
	lines := extractLines(err, c)
	count := 0
	for i, line := range lines {
		if len(line) > 0 {
			count++
			CheckAccessLogFormat(c, line, i)
		}
	}
	return count
}

func checkAccessLogExactValuesOutput(err error, c *check.C, values []accessLogValue) int {
	lines := extractLines(err, c)
	count := 0
	for i, line := range lines {
		fmt.Printf("Line %s", line)
		fmt.Println()
		if len(line) > 0 {
			count++
			if values[i].formatOnly {
				CheckAccessLogFormat(c, line, i)
			} else {
				checkAccessLogExactValues(c, line, i, values[i])
			}
		}
	}
	return count
}

func extractLines(err error, c *check.C) []string {
	accessLog, err := ioutil.ReadFile(traefikTestAccessLogFile)
	c.Assert(err, checker.IsNil)
	lines := strings.Split(string(accessLog), "\n")
	return lines
}

func checkStatsForLogFile(c *check.C) {
	err := try.Do(1*time.Second, func() error {
		if _, errStat := os.Stat(traefikTestLogFile); errStat != nil {
			return fmt.Errorf("could not get stats for log file: %s", errStat)
		}
		return nil
	})
	c.Assert(err, checker.IsNil)
}

func ensureWorkingDirectoryIsClean() {
	os.Remove(traefikTestAccessLogFile)
	os.Remove(traefikTestLogFile)
}

func checkTraefikStarted(c *check.C) []byte {
	traefikLog, err := ioutil.ReadFile(traefikTestLogFile)
	c.Assert(err, checker.IsNil)
	if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
		c.Assert(traefikLog, checker.HasLen, 0)
	}
	return traefikLog
}

func CheckAccessLogFormat(c *check.C, line string, i int) {
	tokens, err := shellwords.Parse(line)
	c.Assert(err, checker.IsNil)
	c.Assert(tokens, checker.HasLen, 14)
	c.Assert(tokens[6], checker.Matches, `^(-|\d{3})$`)
	c.Assert(tokens[10], checker.Equals, fmt.Sprintf("%d", i+1))
	c.Assert(tokens[11], checker.HasPrefix, "frontend")
	c.Assert(tokens[12], checker.HasPrefix, "http://127.0.0.1:808")
	c.Assert(tokens[13], checker.Matches, `^\d+ms$`)
}

func checkAccessLogExactValues(c *check.C, line string, i int, v accessLogValue) {
	tokens, err := shellwords.Parse(line)
	c.Assert(err, checker.IsNil)
	c.Assert(tokens, checker.HasLen, 14)
	c.Assert(tokens[6], checker.Equals, v.code)
	c.Assert(tokens[10], checker.Equals, fmt.Sprintf("%d", i+1))
	c.Assert(tokens[11], checker.HasPrefix, v.value)
	c.Assert(tokens[12], checker.Equals, v.backendName)
	c.Assert(tokens[13], checker.Matches, `^\d+ms$`)
}

func startAccessLogServer(port int) (ts *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Received query %s!\n", r.URL.Path[1:])
	})
	if listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port)); err != nil {
		panic(err)
	} else {
		ts = &httptest.Server{
			Listener: listener,
			Config:   &http.Server{Handler: handler},
		}
		ts.Start()
	}
	return
}

func (s *AccessLogSuite) startContainerWithNameAndLabels(c *check.C, name string, image string, labels map[string]string, args ...string) string {
	container := s.project.StartWithConfig(c, image, d.ContainerConfig{
		Name:   name,
		Cmd:    args,
		Labels: labels,
	})
	waitForTraefik(c, name)

	return container.Name
}

func waitForTraefik(c *check.C, containerName string) {
	// Wait for Traefik to turn ready.
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:7888/api", nil)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains(containerName))
	c.Assert(err, checker.IsNil)
}
