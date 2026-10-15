package integration

import (
	"database/sql"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"gopkg.in/yaml.v3"
)

type ExternalActionSuite struct {
	BaseSuite

	db  *sql.DB
	dsn string
}

func TestExternalActionSuite(t *testing.T) {
	suite.Run(t, new(ExternalActionSuite))
}

func (s *ExternalActionSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.createComposeProject("governance")
	s.composeUp()
	config := mysql.NewConfig()
	config.User = "root"
	config.Passwd = "integration-only"
	config.Net = "tcp"
	config.Addr = net.JoinHostPort(s.getComposeServiceIP("mysql"), "3306")
	config.DBName = "governance"
	config.ParseTime = true
	config.Timeout = time.Second
	s.dsn = config.FormatDSN()
	var err error
	s.db, err = sql.Open("mysql", s.dsn)
	require.NoError(s.T(), err)
	require.NoError(s.T(), try.Do(time.Minute, func() error { return s.db.PingContext(s.T().Context()) }))
}

func (s *ExternalActionSuite) TearDownSuite() {
	if s.db != nil {
		assert.NoError(s.T(), s.db.Close())
	}
	s.BaseSuite.TearDownSuite()
}

func (s *ExternalActionSuite) TestResponseMappingAndPersistentDeduplication() {
	var calls atomic.Int32
	reports := make(chan []byte, 8)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), "Bearer integration-token", r.Header.Get("Authorization"))
		reports <- body
		calls.Add(1)
		_, _ = w.Write([]byte(`{"accepted":true}`))
	}))
	defer target.Close()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":9007199254740993,"state":"done"},{"id":9007199254740994,"state":"running"}]}`))
	}))
	defer backend.Close()
	config := externalactionHTTPConfig(backend.URL, dynamic.Middleware{ExternalAction: &dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{{
			Name: "completed-items", Phase: "after", Mode: "sync", Timeout: ptypes.Duration(5 * time.Second),
			Match:         dynamic.ExternalActionMatch{Methods: []string{http.MethodPost}, PathRegex: "^/events$", ResponseStatuses: []string{"200"}},
			Body:          dynamic.ExternalActionBody{Format: "json", Template: `{"id":{{ jsonPath .Item "$.id" | toJson }},"actor":{{ header . "X-Authenticated-User" | toJson }},"input":{{ jsonPath .Body "$.source" | toJson }}}`},
			ResponseBody:  dynamic.ExternalActionResponseBody{ReadMode: "capture", Format: "json", MaxBytes: 1024},
			ForEach:       &dynamic.ExternalActionForEach{Source: "responseBody", JSONPath: "$.items", MatchJSONPaths: map[string]string{"$.state": "^done$"}},
			Deduplication: &dynamic.ExternalActionDeduplication{KeyTemplate: `item:{{ jsonPath .Item "$.id" }}`, MySQL: &dynamic.ExternalActionDedupMySQL{DSN: s.dsn, Table: "action_dedup", AutoCreate: true}},
			HTTP:          &dynamic.ExternalActionHTTP{URL: target.URL, Auth: &dynamic.ExternalActionAuth{Type: "bearer", Token: "integration-token"}},
		}},
	}})
	cmd := s.startGovernance(config)
	for range 2 {
		body := s.request("/events", `{"source":"example"}`, http.Header{"X-Authenticated-User": {"alice"}})
		assert.Contains(s.T(), string(body), "9007199254740993")
	}
	select {
	case report := <-reports:
		assert.JSONEq(s.T(), `{"id":9007199254740993,"actor":"alice","input":"example"}`, string(report))
		assert.Contains(s.T(), string(report), "9007199254740993")
	case <-time.After(5 * time.Second):
		s.T().Fatal("no action received")
	}
	require.Eventually(s.T(), func() bool {
		var count int
		err := s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM action_dedup WHERE state='completed'").Scan(&count)
		return err == nil && count == 1
	}, 5*time.Second, 100*time.Millisecond)
	s.killCmd(cmd)
	_ = cmd.Wait()
	s.startGovernance(config)
	s.request("/events", `{"source":"example"}`, http.Header{"X-Authenticated-User": {"alice"}})
	assert.EqualValues(s.T(), 1, calls.Load())
}

func (s *ExternalActionSuite) TestActionFailurePolicyAndBodyLimit() {
	var backendCalls atomic.Int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		backendCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer target.Close()
	var rules []dynamic.ExternalActionRule
	for _, policy := range []string{"fail", "ignore"} {
		rules = append(rules, dynamic.ExternalActionRule{
			Name: policy, Phase: "before", Mode: "sync", FailurePolicy: policy,
			Match: dynamic.ExternalActionMatch{PathRegex: "^/" + policy + "$"},
			HTTP:  &dynamic.ExternalActionHTTP{URL: target.URL},
		})
	}
	rules = append(rules, dynamic.ExternalActionRule{
		Name: "bounded", Phase: "before", Mode: "sync", FailurePolicy: "fail",
		Match: dynamic.ExternalActionMatch{PathRegex: "^/bounded$"},
		Body:  dynamic.ExternalActionBody{ReadMode: "eager", MaxBytes: 8, Overflow: "reject", Template: "{{ .RawBody }}"},
		HTTP:  &dynamic.ExternalActionHTTP{URL: target.URL},
	})
	s.startGovernance(externalactionHTTPConfig(backend.URL, dynamic.Middleware{ExternalAction: &dynamic.ExternalAction{Rules: rules}}))
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/fail", http.StatusBadGateway},
		{"/ignore", http.StatusOK},
		{"/bounded", http.StatusRequestEntityTooLarge},
	} {
		req, err := http.NewRequestWithContext(s.T().Context(), http.MethodPost, "http://127.0.0.1:8000"+tc.path, strings.NewReader(`{"message":"too long"}`))
		require.NoError(s.T(), err)
		resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), tc.status, resp.StatusCode)
		_, _ = io.Copy(io.Discard, resp.Body)
		require.NoError(s.T(), resp.Body.Close())
	}
	assert.EqualValues(s.T(), 1, backendCalls.Load())
}

func (s *ExternalActionSuite) TestExecutableReceivesLiteralInput() {
	command, err := exec.LookPath("tee")
	require.NoError(s.T(), err)
	file := filepath.Join(s.T().TempDir(), "adapter-input")
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, r.Body)
	}))
	defer backend.Close()
	s.startGovernance(externalactionHTTPConfig(backend.URL, dynamic.Middleware{ExternalAction: &dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{{
			Name: "adapter", Phase: "before", Mode: "sync", FailurePolicy: "fail",
			Body: dynamic.ExternalActionBody{ReadMode: "eager", Format: "json"},
			Exec: &dynamic.ExternalActionExec{Command: command, Args: []string{file}, Stdin: `{{ jsonPath .Body "$.argument" }}`},
		}},
	}}))
	const input = `{"argument":"literal; $(never-run)"}`
	assert.JSONEq(s.T(), input, string(s.request("/adapter", input, nil)))
	actual, err := os.ReadFile(file)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "literal; $(never-run)", string(actual))
}

func (s *ExternalActionSuite) startGovernance(config dynamic.Configuration) *exec.Cmd {
	s.T().Helper()
	content, err := yaml.Marshal(config)
	require.NoError(s.T(), err)
	file := filepath.Join(s.T().TempDir(), "routing.yml")
	// The file provider renders Go templates before decoding YAML; preserve the
	// action templates for request-time evaluation by emitting the fixture literally.
	literal := "{{ " + strconv.Quote(string(content)) + " }}"
	require.NoError(s.T(), os.WriteFile(file, []byte(literal), 0o600))
	cmd := s.traefikCmd("--providers.file.filename="+file, "--providers.file.watch=false",
		"--entrypoints.web.address=:8000", "--entrypoints.tcp.address=:8090",
		"--entrypoints.udp.address=:8091/udp", "--entrypoints.udp.udp.timeout=1s",
		"--api.insecure=true", "--log.level=DEBUG", "--global.checknewversion=false",
		"--global.sendanonymoususage=false")
	require.NoError(s.T(), try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second,
		try.StatusCodeIs(http.StatusOK), try.BodyContains("governance@file")))
	return cmd
}

func externalactionHTTPConfig(backend string, middleware dynamic.Middleware) dynamic.Configuration {
	return dynamic.Configuration{HTTP: &dynamic.HTTPConfiguration{
		Routers: map[string]*dynamic.Router{"governance": {
			EntryPoints: []string{"web"}, Rule: "PathPrefix(`/`)", Service: "backend", Middlewares: []string{"governance"},
		}},
		Services: map[string]*dynamic.Service{"backend": {
			LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: backend}}},
		}},
		Middlewares: map[string]*dynamic.Middleware{"governance": &middleware},
	}}
}

func (s *ExternalActionSuite) request(path, body string, headers http.Header) []byte {
	s.T().Helper()
	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodPost, "http://127.0.0.1:8000"+path, strings.NewReader(body))
	require.NoError(s.T(), err)
	req.Header = headers.Clone()
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	result, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	return result
}
