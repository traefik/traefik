package integration

import (
	"database/sql"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type AccessGovernanceSuite struct {
	BaseSuite

	db  *sql.DB
	dsn string
}

func TestAccessGovernanceSuite(t *testing.T) {
	suite.Run(t, new(AccessGovernanceSuite))
}

func (s *AccessGovernanceSuite) SetupSuite() {
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

func (s *AccessGovernanceSuite) TearDownSuite() {
	if s.db != nil {
		assert.NoError(s.T(), s.db.Close())
	}
	s.BaseSuite.TearDownSuite()
}

func (s *AccessGovernanceSuite) TestHTTPAuditRoutingAndHotDSNReload() {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.Copy(w, r.Body) }))
	defer backend.Close()
	config := s.auditConfig("audit_http")
	config.Identity = []dynamic.AccessGovernanceIdentitySource{
		{Kind: "header", Name: "X-Authenticated-User"}, {Kind: "cookie", Name: "account"}, {Kind: "json", JSONPath: "$.actor"},
	}
	config.Databases["selected"] = dynamic.AccessGovernanceDatabase{DSN: s.dsn, HTTPTable: "audit_selected", NetworkTable: "audit_selected_network"}
	config.Routes = []dynamic.AccessGovernanceRoute{{Name: "selected", Database: "selected", BodyJSONPaths: map[string]string{"$.destination": "^selected$"}}}
	directory := s.T().TempDir()
	file := filepath.Join(directory, "dsn")
	require.NoError(s.T(), os.Symlink("..data/dsn", file))
	rotate := func(version, material string) {
		s.T().Helper()
		require.NoError(s.T(), os.Mkdir(filepath.Join(directory, version), 0o700))
		require.NoError(s.T(), os.WriteFile(filepath.Join(directory, version, "dsn"), []byte(material), 0o600))
		require.NoError(s.T(), os.Symlink(version, filepath.Join(directory, "..next")))
		require.NoError(s.T(), os.Rename(filepath.Join(directory, "..next"), filepath.Join(directory, "..data")))
	}
	rotate("one", s.dsn)
	destination := config.Databases["default"]
	destination.DSN, destination.DSNFile = "", file
	config.Databases["default"] = destination
	s.startGovernance(accessgovernanceHTTPConfig(backend.URL, dynamic.Middleware{AccessGovernance: config}))
	for _, tc := range []struct {
		body, principal, table string
		headers                http.Header
	}{
		{`{"actor":"ignored"}`, "alice", "audit_http", http.Header{"X-Authenticated-User": {"alice"}}},
		{`{"actor":"ignored"}`, "bob", "audit_http", http.Header{"Cookie": {"account=bob"}}},
		{`{"actor":"carol","destination":"selected"}`, "carol", "audit_selected", nil},
		{`{"actor":9007199254740993}`, "9007199254740993", "audit_http", nil},
		{`{"actor":"x'); DROP TABLE audit_http; --"}`, "x'); DROP TABLE audit_http; --", "audit_http", nil},
	} {
		assert.JSONEq(s.T(), tc.body, string(s.request("/events", tc.body, tc.headers)))
		require.Eventually(s.T(), func() bool {
			var count int
			err := s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM "+tc.table+" WHERE principal_id=? AND status_code=200", tc.principal).Scan(&count)
			return err == nil && count == 1
		}, 10*time.Second, 100*time.Millisecond)
	}
	_, err := s.db.ExecContext(s.T().Context(), "CREATE DATABASE governance_rotated")
	require.NoError(s.T(), err)
	dsn, err := mysql.ParseDSN(s.dsn)
	require.NoError(s.T(), err)
	dsn.DBName = "governance_rotated"
	rotate("two", dsn.FormatDSN())
	// Only the credential file changes; the process and routing configuration stay intact.
	require.Eventually(s.T(), func() bool {
		var count int
		err := s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM governance_rotated.audit_http").Scan(&count)
		return err == nil
	}, 15*time.Second, 100*time.Millisecond)
	require.Eventually(s.T(), func() bool {
		// Schema creation precedes pool activation. Requests during rotation may
		// still use the old pool, so probe until the new destination is active.
		req, requestErr := http.NewRequestWithContext(s.T().Context(), http.MethodPost, "http://127.0.0.1:8000/rotated", nil)
		if requestErr != nil {
			return false
		}
		req.Header.Set("X-Authenticated-User", "dave")
		resp, requestErr := (&http.Client{Timeout: time.Second}).Do(req)
		if requestErr != nil {
			return false
		}
		_ = resp.Body.Close()
		var count int
		err := s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM governance_rotated.audit_http WHERE principal_id='dave'").Scan(&count)
		return err == nil && count >= 1
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *AccessGovernanceSuite) TestTCPAndUDPAudit() {
	tcpServer, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err)
	defer tcpServer.Close()
	go func() {
		conn, acceptErr := tcpServer.Accept()
		if acceptErr == nil {
			defer conn.Close()
			_, _ = io.Copy(conn, conn)
		}
	}()
	udpServer, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(s.T(), err)
	defer udpServer.Close()
	go func() {
		buf := make([]byte, 64)
		n, addr, readErr := udpServer.ReadFrom(buf)
		if readErr == nil {
			_, _ = udpServer.WriteTo(buf[:n], addr)
		}
	}()
	audit := s.auditConfig("audit_transport")
	s.startGovernance(dynamic.Configuration{
		TCP: &dynamic.TCPConfiguration{
			Routers:     map[string]*dynamic.TCPRouter{"governance": {EntryPoints: []string{"tcp"}, Rule: "HostSNI(`*`)", Middlewares: []string{"audit"}, Service: "echo"}},
			Services:    map[string]*dynamic.TCPService{"echo": {LoadBalancer: &dynamic.TCPServersLoadBalancer{Servers: []dynamic.TCPServer{{Address: tcpServer.Addr().String()}}}}},
			Middlewares: map[string]*dynamic.TCPMiddleware{"audit": {AccessGovernance: audit}},
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:     map[string]*dynamic.UDPRouter{"governance": {EntryPoints: []string{"udp"}, Middlewares: []string{"audit"}, Service: "echo"}},
			Services:    map[string]*dynamic.UDPService{"echo": {LoadBalancer: &dynamic.UDPServersLoadBalancer{Servers: []dynamic.UDPServer{{Address: udpServer.LocalAddr().String()}}}}},
			Middlewares: map[string]*dynamic.UDPMiddleware{"audit": {AccessGovernance: audit}},
		},
	})
	for _, tc := range []struct{ protocol, address string }{{"tcp", "127.0.0.1:8090"}, {"udp", "127.0.0.1:8091"}} {
		conn, dialErr := net.DialTimeout(tc.protocol, tc.address, time.Second)
		require.NoError(s.T(), dialErr)
		require.NoError(s.T(), conn.SetDeadline(time.Now().Add(5*time.Second)))
		_, err = conn.Write([]byte("hello"))
		require.NoError(s.T(), err)
		buf := make([]byte, 5)
		_, err = io.ReadFull(conn, buf)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "hello", string(buf))
		require.NoError(s.T(), conn.Close())
		require.Eventually(s.T(), func() bool {
			var received, sent int
			queryErr := s.db.QueryRowContext(s.T().Context(), "SELECT bytes_received, bytes_sent FROM audit_transport_network WHERE protocol=?", tc.protocol).Scan(&received, &sent)
			return queryErr == nil && received == 5 && sent == 5
		}, 10*time.Second, 100*time.Millisecond)
	}
}

func (s *AccessGovernanceSuite) startGovernance(config dynamic.Configuration) {
	s.T().Helper()
	content, err := yaml.Marshal(config)
	require.NoError(s.T(), err)
	file := filepath.Join(s.T().TempDir(), "routing.yml")
	// The file provider renders Go templates before decoding YAML; preserve the
	// action templates for request-time evaluation by emitting the fixture literally.
	literal := "{{ " + strconv.Quote(string(content)) + " }}"
	require.NoError(s.T(), os.WriteFile(file, []byte(literal), 0o600))
	s.traefikCmd("--providers.file.filename="+file, "--providers.file.watch=false",
		"--entrypoints.web.address=:8000", "--entrypoints.tcp.address=:8090",
		"--entrypoints.udp.address=:8091/udp", "--entrypoints.udp.udp.timeout=1s",
		"--api.insecure=true", "--log.level=DEBUG", "--global.checknewversion=false",
		"--global.sendanonymoususage=false")
	require.NoError(s.T(), try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second,
		try.StatusCodeIs(http.StatusOK), try.BodyContains("governance@file")))
}

func accessgovernanceHTTPConfig(backend string, middleware dynamic.Middleware) dynamic.Configuration {
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

func (s *AccessGovernanceSuite) auditConfig(table string) *dynamic.AccessGovernance {
	return &dynamic.AccessGovernance{
		DefaultDatabase: "default",
		Databases: map[string]dynamic.AccessGovernanceDatabase{"default": {
			DSN: s.dsn, HTTPTable: table, NetworkTable: table + "_network",
		}},
		Schema: dynamic.AccessGovernanceSchema{AutoCreate: true},
		Queue:  dynamic.AccessGovernanceQueue{Capacity: 32, Workers: 1, BatchSize: 1, FlushInterval: ptypes.Duration(100 * time.Millisecond)},
	}
}

func (s *AccessGovernanceSuite) request(path, body string, headers http.Header) []byte {
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
