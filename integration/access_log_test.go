package integration

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
)

const (
	traefikTestLogFile       = "traefik.log"
	traefikTestAccessLogFile = "access.log"
)

// AccessLogSuite tests suite.
type AccessLogSuite struct{ BaseSuite }

func TestAccessLogSuite(t *testing.T) {
	suite.Run(t, new(AccessLogSuite))
}

type accessLogValue struct {
	formatOnly bool
	code       string
	user       string
	routerName string
	serviceURL string
}

func (s *AccessLogSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.createComposeProject("access_log")
	s.composeUp()
}

func (s *AccessLogSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *AccessLogSuite) TearDownTest() {
	s.displayTraefikLogFile(traefikTestLogFile)
	_ = os.Remove(traefikTestAccessLogFile)
}

func (s *AccessLogSuite) TestAccessLog() {
	ensureWorkingDirectoryIsClean()

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	defer func() {
		traefikLog, err := os.ReadFile(traefikTestLogFile)
		require.NoError(s.T(), err)
		log.Info().Msg(string(traefikLog))
	}()

	s.waitForTraefik("server1")

	s.checkStatsForLogFile()

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Make some requests
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "frontend1.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "frontend2.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogOutput()

	assert.Equal(s.T(), 3, count)

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogAuthFrontend() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: false,
			code:       "401",
			user:       "-",
			routerName: "rt-authFrontend",
			serviceURL: "-",
		},
		{
			formatOnly: false,
			code:       "401",
			user:       "test",
			routerName: "rt-authFrontend",
			serviceURL: "-",
		},
		{
			formatOnly: false,
			code:       "200",
			user:       "test",
			routerName: "rt-authFrontend",
			serviceURL: "http://172.31.42",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("authFrontend")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8006/", nil)
	require.NoError(s.T(), err)
	req.Host = "frontend.auth.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	require.NoError(s.T(), err)

	req.SetBasicAuth("test", "")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	require.NoError(s.T(), err)

	req.SetBasicAuth("test", "test")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogDigestAuthMiddleware() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: false,
			code:       "401",
			user:       "-",
			routerName: "rt-digestAuthMiddleware",
			serviceURL: "-",
		},
		{
			formatOnly: false,
			code:       "401",
			user:       "test",
			routerName: "rt-digestAuthMiddleware",
			serviceURL: "-",
		},
		{
			formatOnly: false,
			code:       "200",
			user:       "test",
			routerName: "rt-digestAuthMiddleware",
			serviceURL: "http://172.31.42",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("digestAuthMiddleware")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8008/", nil)
	require.NoError(s.T(), err)
	req.Host = "entrypoint.digest.auth.docker.local"

	resp, err := try.ResponseUntilStatusCode(req, 500*time.Millisecond, http.StatusUnauthorized)
	require.NoError(s.T(), err)

	digest := digestParts(resp)
	digest["uri"] = "/"
	digest["method"] = http.MethodGet
	digest["username"] = "test"
	digest["password"] = "wrong"

	req.Header.Set("Authorization", getDigestAuthorization(digest))
	req.Header.Set("Content-Type", "application/json")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	require.NoError(s.T(), err)

	digest["password"] = "test"

	req.Header.Set("Authorization", getDigestAuthorization(digest))

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
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
	if _, err := digest.Write([]byte(data)); err != nil {
		log.Error().Err(err).Send()
	}

	return hex.EncodeToString(digest.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		log.Error().Err(err).Send()
	}

	return hex.EncodeToString(b)[:16]
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

func (s *AccessLogSuite) TestAccessLogFrontendRedirect() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: false,
			code:       "302",
			user:       "-",
			routerName: "rt-frontendRedirect",
			serviceURL: "-",
		},
		{
			formatOnly: true,
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("frontendRedirect")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test frontend redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8005/test", nil)
	require.NoError(s.T(), err)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogJSONFrontendRedirect() {
	ensureWorkingDirectoryIsClean()

	type logLine struct {
		DownstreamStatus int    `json:"downstreamStatus"`
		OriginStatus     int    `json:"originStatus"`
		RouterName       string `json:"routerName"`
		ServiceName      string `json:"serviceName"`
	}

	expected := []logLine{
		{
			DownstreamStatus: 302,
			OriginStatus:     0,
			RouterName:       "rt-frontendRedirect@docker",
			ServiceName:      "",
		},
		{
			DownstreamStatus: 200,
			OriginStatus:     200,
			RouterName:       "rt-server0@docker",
			ServiceName:      "service1@docker",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log_json_config.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("frontendRedirect")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test frontend redirect
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8005/test", nil)
	require.NoError(s.T(), err)
	req.Host = ""

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	lines := s.extractLines()
	assert.GreaterOrEqual(s.T(), len(lines), len(expected))

	for i, line := range lines {
		if line == "" {
			continue
		}
		var logline logLine
		err := json.Unmarshal([]byte(line), &logline)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), expected[i].DownstreamStatus, logline.DownstreamStatus)
		assert.Equal(s.T(), expected[i].OriginStatus, logline.OriginStatus)
		assert.Equal(s.T(), expected[i].RouterName, logline.RouterName)
		assert.Equal(s.T(), expected[i].ServiceName, logline.ServiceName)
	}
}

func (s *AccessLogSuite) TestAccessLogRateLimit() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: true,
		},
		{
			formatOnly: true,
		},
		{
			formatOnly: false,
			code:       "429",
			user:       "-",
			routerName: "rt-rateLimit",
			serviceURL: "-",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("rateLimit")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8007/", nil)
	require.NoError(s.T(), err)
	req.Host = "ratelimit.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogBackendNotFound() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: false,
			code:       "404",
			user:       "-",
			routerName: "-",
			serviceURL: "-",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.waitForTraefik("server1")

	s.checkStatsForLogFile()

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "backendnotfound.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogFrontendAllowlist() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: false,
			code:       "403",
			user:       "-",
			routerName: "rt-frontendAllowlist",
			serviceURL: "-",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("frontendAllowlist")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test rate limit
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "frontend.allowlist.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusForbidden), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogAuthFrontendSuccess() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: false,
			code:       "200",
			user:       "test",
			routerName: "rt-authFrontend",
			serviceURL: "http://172.31.42",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("authFrontend")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test auth entrypoint
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8006/", nil)
	require.NoError(s.T(), err)
	req.Host = "frontend.auth.docker.local"
	req.SetBasicAuth("test", "test")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogPreflightHeadersMiddleware() {
	ensureWorkingDirectoryIsClean()

	expected := []accessLogValue{
		{
			formatOnly: false,
			code:       "200",
			user:       "-",
			routerName: "rt-preflightCORS",
			serviceURL: "-",
		},
	}

	// Start Traefik
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	s.checkStatsForLogFile()

	s.waitForTraefik("preflightCORS")

	// Verify Traefik started OK
	s.checkTraefikStarted()

	// Test preflight response
	req, err := http.NewRequest(http.MethodOptions, "http://127.0.0.1:8009/", nil)
	require.NoError(s.T(), err)
	req.Host = "preflight.docker.local"
	req.Header.Set("Origin", "whatever")
	req.Header.Set("Access-Control-Request-Method", "GET")

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Verify access.log output as expected
	count := s.checkAccessLogExactValuesOutput(expected)

	assert.GreaterOrEqual(s.T(), count, len(expected))

	// Verify no other Traefik problems
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) TestAccessLogDisabledForInternals() {
	ensureWorkingDirectoryIsClean()

	// Start Traefik.
	s.traefikCmd(withConfigFile("fixtures/access_log/access_log_base.toml"))

	defer func() {
		traefikLog, err := os.ReadFile(traefikTestLogFile)
		require.NoError(s.T(), err)
		log.Info().Msg(string(traefikLog))
	}()

	// waitForTraefik makes at least one call to the rawdata api endpoint,
	// but the logs for this endpoint are ignored in checkAccessLogOutput.
	s.waitForTraefik("service3")

	s.checkStatsForLogFile()

	// Verify Traefik started OK.
	s.checkTraefikStarted()

	// Make some requests on the internal ping router.
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/ping", nil)
	require.NoError(s.T(), err)

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	// Make some requests on the custom ping router.
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8010/ping", nil)
	require.NoError(s.T(), err)
	req.Host = "ping.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)
	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	require.NoError(s.T(), err)

	// Verify access.log output as expected.
	count := s.checkAccessLogOutput()

	require.Equal(s.T(), 0, count)

	// Verify no other Traefik problems.
	s.checkNoOtherTraefikProblems()
}

func (s *AccessLogSuite) checkNoOtherTraefikProblems() {
	traefikLog, err := os.ReadFile(traefikTestLogFile)
	require.NoError(s.T(), err)
	if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
	}
}

func (s *AccessLogSuite) checkAccessLogOutput() int {
	s.T().Helper()

	lines := s.extractLines()
	count := 0
	for i, line := range lines {
		if len(line) > 0 {
			count++
			s.CheckAccessLogFormat(line, i)
		}
	}
	return count
}

func (s *AccessLogSuite) checkAccessLogExactValuesOutput(values []accessLogValue) int {
	s.T().Helper()

	lines := s.extractLines()
	count := 0
	for i, line := range lines {
		fmt.Println(line)
		if len(line) > 0 {
			count++
			if values[i].formatOnly {
				s.CheckAccessLogFormat(line, i)
			} else {
				s.checkAccessLogExactValues(line, i, values[i])
			}
		}
	}
	return count
}

func (s *AccessLogSuite) extractLines() []string {
	s.T().Helper()

	accessLog, err := os.ReadFile(traefikTestAccessLogFile)
	require.NoError(s.T(), err)

	lines := strings.Split(string(accessLog), "\n")

	var clean []string
	for _, line := range lines {
		if !strings.Contains(line, "/api/rawdata") {
			clean = append(clean, line)
		}
	}
	return clean
}

func (s *AccessLogSuite) checkStatsForLogFile() {
	s.T().Helper()

	err := try.Do(1*time.Second, func() error {
		if _, errStat := os.Stat(traefikTestLogFile); errStat != nil {
			return fmt.Errorf("could not get stats for log file: %w", errStat)
		}
		return nil
	})
	require.NoError(s.T(), err)
}

func ensureWorkingDirectoryIsClean() {
	os.Remove(traefikTestAccessLogFile)
	os.Remove(traefikTestLogFile)
}

func (s *AccessLogSuite) checkTraefikStarted() []byte {
	s.T().Helper()

	traefikLog, err := os.ReadFile(traefikTestLogFile)
	require.NoError(s.T(), err)
	if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
	}
	return traefikLog
}

func (s *BaseSuite) CheckAccessLogFormat(line string, i int) {
	s.T().Helper()

	results, err := accesslog.ParseAccessLog(line)
	require.NoError(s.T(), err)
	assert.Len(s.T(), results, 14)
	assert.Regexp(s.T(), `^(-|\d{3})$`, results[accesslog.OriginStatus])
	count, _ := strconv.Atoi(results[accesslog.RequestCount])
	assert.GreaterOrEqual(s.T(), count, i+1)
	assert.Regexp(s.T(), `"(rt-.+@docker|api@internal)"`, results[accesslog.RouterName])
	assert.True(s.T(), strings.HasPrefix(results[accesslog.ServiceURL], `"http://`))
	assert.Regexp(s.T(), `^\d+ms$`, results[accesslog.Duration])
}

func (s *AccessLogSuite) checkAccessLogExactValues(line string, i int, v accessLogValue) {
	s.T().Helper()

	results, err := accesslog.ParseAccessLog(line)
	require.NoError(s.T(), err)
	assert.Len(s.T(), results, 14)
	if len(v.user) > 0 {
		assert.Equal(s.T(), v.user, results[accesslog.ClientUsername])
	}
	assert.Equal(s.T(), v.code, results[accesslog.OriginStatus])
	count, _ := strconv.Atoi(results[accesslog.RequestCount])
	assert.GreaterOrEqual(s.T(), count, i+1)
	assert.Regexp(s.T(), `^"?`+v.routerName+`.*(@docker)?$`, results[accesslog.RouterName])
	assert.Regexp(s.T(), `^"?`+v.serviceURL+`.*$`, results[accesslog.ServiceURL])
	assert.Regexp(s.T(), `^\d+ms$`, results[accesslog.Duration])
}
