package integration

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

const (
	traefikTestLogFile       = "traefik.log"
	traefikTestAccessLogFile = "access.log"
)

// AccessLogSuite
type AccessLogSuite struct{ BaseSuite }

type accessLogValue struct {
	formatOnly   bool
	code         string
	user         string
	frontendName string
	backendURL   string
}

func (s *AccessLogSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "access_log")
	s.composeProject.Start(c)

	s.composeProject.Container(c, "server0")
	s.composeProject.Container(c, "server1")
	s.composeProject.Container(c, "server2")
	s.composeProject.Container(c, "server3")
}

func (s *AccessLogSuite) TearDownTest(c *check.C) {
	displayTraefikLogFile(c, traefikTestLogFile)
	os.Remove(traefikTestAccessLogFile)
}

func (s *AccessLogSuite) TestAccessLog(c *check.C) {
	ensureWorkingDirectoryIsClean()

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	waitForTraefik(c, "server1")

	checkStatsForLogFile(c)

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Make some requests
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend1.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend2.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogOutput(c)

	c.Assert(count, checker.GreaterOrEqualThan, 3)

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogAuthFrontend(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "401",
			user:         "-",
			frontendName: "Auth for frontend-Host-frontend-auth-docker-local",
			backendURL:   "/",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "authFrontend")

	waitForTraefik(c, "authFrontend")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test auth frontend
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8006/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend.auth.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogAuthEntrypoint(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "401",
			user:         "-",
			frontendName: "Auth for entrypoint",
			backendURL:   "/",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "authEntrypoint")

	waitForTraefik(c, "authEntrypoint")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8004/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.auth.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogAuthEntrypointSuccess(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "200",
			user:         "test",
			frontendName: "Host-entrypoint-auth-docker",
			backendURL:   "http://172.17.0",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "authEntrypoint")

	waitForTraefik(c, "authEntrypoint")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8004/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.auth.docker.local"
	req.SetBasicAuth("test", "test")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogDigestAuthEntrypoint(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "401",
			user:         "-",
			frontendName: "Auth for entrypoint",
			backendURL:   "/",
		},
		{
			formatOnly:   false,
			code:         "200",
			user:         "test",
			frontendName: "Host-entrypoint-digest-auth-docker",
			backendURL:   "http://172.17.0",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "digestAuthEntrypoint")

	waitForTraefik(c, "digestAuthEntrypoint")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8008/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.digest.auth.docker.local"

	resp, err := try.ResponseUntilStatusCode(req, 500*time.Millisecond, http.StatusUnauthorized)
	c.Assert(err, checker.IsNil)

	digestParts := digestParts(resp)
	digestParts["uri"] = "/"
	digestParts["method"] = http.MethodGet
	digestParts["username"] = "test"
	digestParts["password"] = "test"

	req.Header.Set("Authorization", getDigestAuthorization(digestParts))
	req.Header.Set("Content-Type", "application/json")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

// Thanks to mvndaai for digest authentication
// https://stackoverflow.com/questions/39474284/how-do-you-do-a-http-post-with-digest-authentication-in-golang/39481441#39481441
func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop", "opaque"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	return result
}

func getMD5(data string) string {
	digest := md5.New()
	digest.Write([]byte(data))
	return fmt.Sprintf("%x", digest.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthorization(digestParts map[string]string) string {
	d := digestParts
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	nonceCount := "00000001"
	cnonce := getCnonce()
	response := getMD5(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc=%s, qop=%s, response="%s", opaque="%s", algorithm="MD5"`,
		d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response, d["opaque"])
	return authorization
}

func (s *AccessLogSuite) TestAccessLogEntrypointRedirect(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "302",
			user:         "-",
			frontendName: "entrypoint redirect for httpRedirect",
			backendURL:   "/",
		},
		{
			formatOnly: true,
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "entrypointRedirect")

	waitForTraefik(c, "entrypointRedirect")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test entrypoint redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8001/test", nil)
	c.Assert(err, checker.IsNil)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogFrontendRedirect(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "302",
			user:         "-",
			frontendName: "frontend redirect for frontend-Path-",
			backendURL:   "/",
		},
		{
			formatOnly: true,
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "frontendRedirect")

	waitForTraefik(c, "frontendRedirect")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test frontend redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8005/test", nil)
	c.Assert(err, checker.IsNil)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogRateLimit(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: true,
		},
		{
			formatOnly: true,
		},
		{
			formatOnly:   false,
			code:         "429",
			user:         "-",
			frontendName: "rate limit for frontend-Host-ratelimit",
			backendURL:   "/",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "rateLimit")

	waitForTraefik(c, "rateLimit")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8007/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "ratelimit.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogBackendNotFound(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "404",
			user:         "-",
			frontendName: "backend not found",
			backendURL:   "/",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	waitForTraefik(c, "server1")

	checkStatsForLogFile(c)

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "backendnotfound.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogEntrypointWhitelist(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "403",
			user:         "-",
			frontendName: "ipwhitelister for entrypoint httpWhitelistReject",
			backendURL:   "/",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "entrypointWhitelist")

	waitForTraefik(c, "entrypointWhitelist")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8002/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.whitelist.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusForbidden), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogFrontendWhitelist(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "403",
			user:         "-",
			frontendName: "ipwhitelister for frontend-Host-frontend-whitelist",
			backendURL:   "/",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "frontendWhitelist")

	waitForTraefik(c, "frontendWhitelist")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend.whitelist.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusForbidden), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func (s *AccessLogSuite) TestAccessLogAuthFrontendSuccess(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:   false,
			code:         "200",
			user:         "test",
			frontendName: "Host-frontend-auth-docker",
			backendURL:   "http://172.17.0",
		},
	}

	// Start Traefik
	cmd, display := s.traefikCmd(withConfigFile("fixtures/access_log_config.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "authFrontend")

	waitForTraefik(c, "authFrontend")

	// Verify Traefik started OK
	checkTraefikStarted(c)

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8006/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend.auth.docker.local"
	req.SetBasicAuth("test", "test")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(c)
}

func checkNoOtherTraefikProblems(c *check.C) {
	traefikLog, err := ioutil.ReadFile(traefikTestLogFile)
	c.Assert(err, checker.IsNil)
	if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
		c.Assert(traefikLog, checker.HasLen, 0)
	}
}

func checkAccessLogOutput(c *check.C) int {
	lines := extractLines(c)
	count := 0
	for i, line := range lines {
		if len(line) > 0 {
			count++
			CheckAccessLogFormat(c, line, i)
		}
	}
	return count
}

func checkAccessLogExactValuesOutput(c *check.C, values []accessLogValue) int {
	lines := extractLines(c)
	count := 0
	for i, line := range lines {
		fmt.Printf(line)
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

func extractLines(c *check.C) []string {
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
	results, err := accesslog.ParseAccessLog(line)
	c.Assert(err, checker.IsNil)
	c.Assert(results, checker.HasLen, 14)
	c.Assert(results[accesslog.OriginStatus], checker.Matches, `^(-|\d{3})$`)
	c.Assert(results[accesslog.RequestCount], checker.Equals, fmt.Sprintf("%d", i+1))
	c.Assert(results[accesslog.FrontendName], checker.HasPrefix, "\"Host-")
	c.Assert(results[accesslog.BackendURL], checker.HasPrefix, "\"http://")
	c.Assert(results[accesslog.Duration], checker.Matches, `^\d+ms$`)
}

func checkAccessLogExactValues(c *check.C, line string, i int, v accessLogValue) {
	results, err := accesslog.ParseAccessLog(line)
	c.Assert(err, checker.IsNil)
	c.Assert(results, checker.HasLen, 14)
	if len(v.user) > 0 {
		c.Assert(results[accesslog.ClientUsername], checker.Equals, v.user)
	}
	c.Assert(results[accesslog.OriginStatus], checker.Equals, v.code)
	c.Assert(results[accesslog.RequestCount], checker.Equals, fmt.Sprintf("%d", i+1))
	c.Assert(results[accesslog.FrontendName], checker.Matches, `^"?`+v.frontendName+`.*$`)
	c.Assert(results[accesslog.BackendURL], checker.Matches, `^"?`+v.backendURL+`.*$`)
	c.Assert(results[accesslog.Duration], checker.Matches, `^\d+ms$`)
}

func waitForTraefik(c *check.C, containerName string) {
	// Wait for Traefik to turn ready.
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api", nil)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains(containerName))
	c.Assert(err, checker.IsNil)
}

func displayTraefikLogFile(c *check.C, path string) {
	if c.Failed() {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			content, errRead := ioutil.ReadFile(path)
			fmt.Printf("%s: Traefik logs: \n", c.TestName())
			if errRead == nil {
				fmt.Println(content)
			} else {
				fmt.Println(errRead)
			}
		} else {
			fmt.Printf("%s: No Traefik logs.\n", c.TestName())
		}
		errRemove := os.Remove(path)
		if errRemove != nil {
			fmt.Println(errRemove)
		}
	}
}
