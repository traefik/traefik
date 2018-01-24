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
	"github.com/go-check/check"
	"github.com/mattn/go-shellwords"
	checker "github.com/vdemeester/shakers"
)

const (
	traefikTestLogFile       = "traefik.log"
	traefikTestAccessLogFile = "access.log"
)

// AccessLogSuite
type AccessLogSuite struct{ BaseSuite }

type accessLogValue struct {
	formatOnly  bool
	code        string
	user        string
	value       string
	backendName string
}

func (s *AccessLogSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "access_log")
	s.composeProject.Start(c)

	s.composeProject.Container(c, "server0")
	s.composeProject.Container(c, "server1")
	s.composeProject.Container(c, "server2")
	s.composeProject.Container(c, "server3")
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

	waitForTraefik(c, "server1")

	checkStatsForLogFile(c)

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

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
	count := checkAccessLogOutput(err, c)

	c.Assert(count, checker.GreaterOrEqualThan, 3)

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogAuthFrontend(c *check.C) {
	// Ensure working directory is clean
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "401",
			user:        "-",
			value:       "Auth for frontend-Host-frontend-auth-docker-local",
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

	s.composeProject.Container(c, "authFrontend")

	waitForTraefik(c, "authFrontend")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test auth frontend
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8006/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend.auth.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogAuthEntrypoint(c *check.C) {
	// Ensure working directory is clean
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "401",
			user:        "-",
			value:       "Auth for entrypoint",
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

	s.composeProject.Container(c, "authEntrypoint")

	waitForTraefik(c, "authEntrypoint")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8004/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.auth.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogAuthEntrypointSuccess(c *check.C) {
	// Ensure working directory is clean
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "200",
			user:        "test",
			value:       "Host-entrypoint-auth-docker",
			backendName: "http://172.17.0",
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

	s.composeProject.Container(c, "authEntrypoint")

	waitForTraefik(c, "authEntrypoint")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8004/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.auth.docker.local"
	req.SetBasicAuth("test", "test")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogDigestAuthEntrypoint(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "401",
			user:        "-",
			value:       "Auth for entrypoint",
			backendName: "-",
		},
		{
			formatOnly:  false,
			code:        "200",
			user:        "test",
			value:       "Host-entrypoint-digest-auth-docker",
			backendName: "http://172.17.0",
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

	s.composeProject.Container(c, "digestAuthEntrypoint")

	waitForTraefik(c, "digestAuthEntrypoint")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

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
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
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
			formatOnly:  false,
			code:        "302",
			user:        "-",
			value:       "entrypoint redirect for frontend-",
			backendName: "-",
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

	defer os.Remove(traefikTestAccessLogFile)
	defer os.Remove(traefikTestLogFile)

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "entrypointRedirect")

	waitForTraefik(c, "entrypointRedirect")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test entrypoint redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8001/test", nil)
	c.Assert(err, checker.IsNil)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogFrontendRedirect(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "302",
			user:        "-",
			value:       "frontend redirect for frontend-Path-",
			backendName: "-",
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

	defer os.Remove(traefikTestAccessLogFile)
	defer os.Remove(traefikTestLogFile)

	checkStatsForLogFile(c)

	s.composeProject.Container(c, "frontendRedirect")

	waitForTraefik(c, "frontendRedirect")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test frontend redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8005/test", nil)
	c.Assert(err, checker.IsNil)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
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
			formatOnly:  false,
			code:        "429",
			user:        "-",
			value:       "rate limit for frontend-Host-ratelimit",
			backendName: "/",
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

	s.composeProject.Container(c, "rateLimit")

	waitForTraefik(c, "rateLimit")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

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
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogBackendNotFound(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "404",
			user:        "-",
			value:       "backend not found",
			backendName: "/",
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

	waitForTraefik(c, "server1")

	checkStatsForLogFile(c)

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "backendnotfound.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogEntrypointWhitelist(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "403",
			user:        "-",
			value:       "ipwhitelister for entrypoint httpWhitelistReject",
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

	s.composeProject.Container(c, "entrypointWhitelist")

	waitForTraefik(c, "entrypointWhitelist")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8002/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "entrypoint.whitelist.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusForbidden), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
}

func (s *AccessLogSuite) TestAccessLogFrontendWhitelist(c *check.C) {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly:  false,
			code:        "403",
			user:        "-",
			value:       "ipwhitelister for frontend-Host-frontend-whitelist",
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

	s.composeProject.Container(c, "frontendWhitelist")

	waitForTraefik(c, "frontendWhitelist")

	// Verify Traefik started OK
	traefikLog := checkTraefikStarted(c)

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend.whitelist.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusForbidden), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	count := checkAccessLogExactValuesOutput(err, c, expected)

	c.Assert(count, checker.GreaterOrEqualThan, len(expected))

	// Verify no other Traefik problems
	checkNoOtherTraefikProblems(traefikLog, err, c)
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
	c.Assert(tokens[11], checker.HasPrefix, "Host-")
	c.Assert(tokens[12], checker.HasPrefix, "http://")
	c.Assert(tokens[13], checker.Matches, `^\d+ms$`)
}

func checkAccessLogExactValues(c *check.C, line string, i int, v accessLogValue) {
	tokens, err := shellwords.Parse(line)
	c.Assert(err, checker.IsNil)
	c.Assert(tokens, checker.HasLen, 14)
	if len(v.user) > 0 {
		c.Assert(tokens[2], checker.Equals, v.user)
	}
	c.Assert(tokens[6], checker.Equals, v.code)
	c.Assert(tokens[10], checker.Equals, fmt.Sprintf("%d", i+1))
	c.Assert(tokens[11], checker.HasPrefix, v.value)
	c.Assert(tokens[12], checker.HasPrefix, v.backendName)
	c.Assert(tokens[13], checker.Matches, `^\d+ms$`)
}

func waitForTraefik(c *check.C, containerName string) {
	// Wait for Traefik to turn ready.
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api", nil)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains(containerName))
	c.Assert(err, checker.IsNil)
}
