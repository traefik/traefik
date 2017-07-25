package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/scroll"
	"github.com/vulcand/vulcand/api"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/engine/memng"
	"github.com/vulcand/vulcand/plugin/registry"
	"github.com/vulcand/vulcand/proxy"
	"github.com/vulcand/vulcand/secret"
	"github.com/vulcand/vulcand/stapler"
	"github.com/vulcand/vulcand/supervisor"
	"github.com/vulcand/vulcand/testutils"

	. "gopkg.in/check.v1"
)

const OK = ".*OK.*"

func TestVulcanCommandLineTool(t *testing.T) { TestingT(t) }

type CmdSuite struct {
	ng         engine.Engine
	out        *bytes.Buffer
	cmd        *Command
	testServer *httptest.Server
	sv         *supervisor.Supervisor
}

var _ = Suite(&CmdSuite{})

func (s *CmdSuite) SetUpTest(c *C) {
	s.ng = memng.New(registry.GetRegistry())

	newProxy := func(id int) (proxy.Proxy, error) {
		return proxy.New(id, stapler.New(), proxy.Options{})
	}

	sv := supervisor.New(newProxy, s.ng, make(chan error), supervisor.Options{})
	sv.Start()
	s.sv = sv

	app := scroll.NewApp()
	api.InitProxyController(s.ng, sv, app)
	s.testServer = httptest.NewServer(app.GetHandler())

	s.out = &bytes.Buffer{}
	s.cmd = &Command{registry: registry.GetRegistry(), out: s.out, vulcanUrl: s.testServer.URL}
}

func (s *CmdSuite) TearDownTest(c *C) {
	s.sv.Stop(true)
}

func (s *CmdSuite) runString(in string) string {
	return s.run(strings.Split(in, " ")...)
}

func (s *CmdSuite) run(params ...string) string {
	args := []string{"vctl"}
	args = append(args, params...)
	args = append(args, fmt.Sprintf("--vulcan=%s", s.testServer.URL))
	s.out = &bytes.Buffer{}
	s.cmd = &Command{registry: registry.GetRegistry(), out: s.out, vulcanUrl: s.testServer.URL}
	s.cmd.Run(args)
	return strings.Replace(s.out.String(), "\n", " ", -1)
}

func (s *CmdSuite) TestStatus(c *C) {
	c.Assert(s.run("top", "--refresh", "0"), Matches, ".*Frontend.*")
}

func (s *CmdSuite) TestHostCRUD(c *C) {
	host := "localhost"
	c.Assert(s.run("host", "upsert", "-name", host), Matches, OK)

	keyPair := testutils.NewTestKeyPair()

	fKey, err := ioutil.TempFile("", "vulcand")
	c.Assert(err, IsNil)
	defer fKey.Close()
	fKey.Write(keyPair.Key)

	fCert, err := ioutil.TempFile("", "vulcand")
	c.Assert(err, IsNil)
	defer fCert.Close()
	fCert.Write(keyPair.Cert)

	c.Assert(s.run("host", "upsert", "-name", host,
		"-privateKey", fKey.Name(), "-cert", fCert.Name(),
		"-ocsp", "-ocspPeriod", "2h", "-ocspResponder", "http://a.com", "-ocspResponder", "http://b.com", "-ocspSkipCheck"), Matches, OK)

	h, err := s.ng.GetHost(engine.HostKey{Name: host})
	c.Assert(err, IsNil)
	c.Assert(h.Settings.KeyPair, DeepEquals, keyPair)

	c.Assert(h.Settings.OCSP.Enabled, Equals, true)
	c.Assert(h.Settings.OCSP.Period, Equals, "2h0m0s")
	c.Assert(h.Settings.OCSP.Responders, DeepEquals, []string{"http://a.com", "http://b.com"})
	c.Assert(h.Settings.OCSP.SkipSignatureCheck, Equals, true)

	c.Assert(s.run("host", "show", "-name", host), Matches, ".*"+host+".*")
	c.Assert(s.run("host", "rm", "-name", host), Matches, OK)
}

func (s *CmdSuite) TestLogSeverity(c *C) {
	for _, sev := range []log.Level{log.InfoLevel, log.WarnLevel, log.ErrorLevel} {
		c.Assert(s.run("log", "set_severity", "-s", sev.String()), Matches, ".*updated.*")
		c.Assert(s.run("log", "get_severity"), Matches, fmt.Sprintf(".*%v.*", sev))
	}
}

func (s *CmdSuite) TestListenerCRUD(c *C) {
	host := "host"
	c.Assert(s.run("host", "upsert", "-name", host), Matches, OK)
	l := "l1"
	c.Assert(s.run("listener", "upsert", "-id", l, "-proto", "http", "-addr", "localhost:11300"), Matches, OK)
	c.Assert(s.run("listener", "ls"), Matches, fmt.Sprintf(".*%v.*", "http"))
	c.Assert(s.run("listener", "show", "-id", l), Matches, fmt.Sprintf(".*%v.*", "http"))
	c.Assert(s.run("listener", "rm", "-id", l), Matches, OK)

	c.Assert(s.run("listener", "upsert", "-id", l, "-proto", "http", "-addr", "localhost:11300", "-scope", `Host("localhost")`), Matches, OK)
}

func (s *CmdSuite) TestHTTPSListenerCRUD(c *C) {
	host := "host"
	c.Assert(s.run("host", "upsert", "-name", host), Matches, OK)
	l := "l1"
	c.Assert(
		s.run("listener", "upsert", "-id", l, "-proto", "https", "-addr", "localhost:11300",
			// TLS parameters
			"-tlsSkipVerify", "-tlsPreferServerCS", "-tlsSessionTicketsOff",
			"-tlsMinV=VersionTLS11", "-tlsMaxV=VersionTLS12",
			"-tlsCS=TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
			"-tlsCS=TLS_ECDHE_RSA_WITH_RC4_128_SHA"), Matches, OK)

	c.Assert(s.run("listener", "rm", "-id", l), Matches, OK)
}

func (s *CmdSuite) TestBackendCRUD(c *C) {
	b := "bk1"
	c.Assert(s.run("backend", "upsert", "-id", b), Matches, OK)
	c.Assert(s.run("backend", "ls"), Matches, fmt.Sprintf(".*%s.*", b))

	c.Assert(s.run(
		"backend", "upsert",
		"-id", b,
		// Timeouts
		"-readTimeout", "1s", "-dialTimeout", "2s", "-handshakeTimeout", "3s",
		// Keep Alive parameters
		"-keepAlivePeriod", "4s", "-maxIdleConns", "5",
		// TLS parameters
		"-tlsSkipVerify", "-tlsPreferServerCS", "-tlsSessionTicketsOff",
		"-tlsMinV=VersionTLS11", "-tlsMaxV=VersionTLS12",
		"-tlsCS=TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		"-tlsCS=TLS_ECDHE_RSA_WITH_RC4_128_SHA",
	),
		Matches, OK)

	val, err := s.ng.GetBackend(engine.BackendKey{Id: b})
	c.Assert(err, IsNil)
	o := val.HTTPSettings()

	c.Assert(o.Timeouts.Read, Equals, "1s")
	c.Assert(o.Timeouts.Dial, Equals, "2s")
	c.Assert(o.Timeouts.TLSHandshake, Equals, "3s")

	c.Assert(o.KeepAlive.Period, Equals, "4s")
	c.Assert(o.KeepAlive.MaxIdleConnsPerHost, Equals, 5)

	c.Assert(o.TLS.InsecureSkipVerify, Equals, true)
	c.Assert(o.TLS.PreferServerCipherSuites, Equals, true)
	c.Assert(o.TLS.SessionTicketsDisabled, Equals, true)
	c.Assert(o.TLS.MinVersion, Equals, "VersionTLS11")
	c.Assert(o.TLS.MaxVersion, Equals, "VersionTLS12")
	c.Assert(o.TLS.SessionTicketsDisabled, Equals, true)
	c.Assert(o.TLS.CipherSuites, DeepEquals,
		[]string{
			"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
			"TLS_ECDHE_RSA_WITH_RC4_128_SHA",
		})
	c.Assert(s.run("backend", "rm", "-id", b), Matches, OK)
}

func (s *CmdSuite) TestBackendSessionCacheCRUD(c *C) {
	b := "bk1"
	c.Assert(s.run("backend", "upsert", "-id", b), Matches, OK)
	c.Assert(s.run("backend", "ls"), Matches, fmt.Sprintf(".*%s.*", b))

	c.Assert(s.run(
		"backend", "upsert",
		"-id", b,
		// Timeouts
		"-readTimeout", "1s", "-dialTimeout", "2s", "-handshakeTimeout", "3s",
		// Keep Alive parameters
		"-keepAlivePeriod", "4s", "-maxIdleConns", "5",
		// TLS parameters
		"-tlsSessionCache=LRU",
		"-tlsSessionCacheCapacity=1023",
	),
		Matches, OK)

	val, err := s.ng.GetBackend(engine.BackendKey{Id: b})
	c.Assert(err, IsNil)
	o := val.HTTPSettings()

	c.Assert(o.TLS.SessionCache.Type, Equals, "LRU")
	c.Assert(o.TLS.SessionCache.Settings.Capacity, Equals, 1023)
	c.Assert(s.run("backend", "rm", "-id", b), Matches, OK)
}

func (s *CmdSuite) TestServerCRUD(c *C) {
	b := "bk1"
	c.Assert(s.run("backend", "upsert", "-id", b), Matches, OK)
	srv := "srv1"
	c.Assert(s.run("server", "upsert", "-id", srv, "-url", "http://localhost:5000", "-b", b), Matches, OK)

	c.Assert(s.run("server", "ls", "-b", b), Matches, fmt.Sprintf(".*%v.*", "http://localhost:5000"))
	c.Assert(s.run("server", "show", "-id", srv, "-b", b), Matches, fmt.Sprintf(".*%v.*", "http://localhost:5000"))

	c.Assert(s.run("server", "rm", "-id", srv, "-b", b), Matches, OK)
	c.Assert(s.run("backend", "rm", "-id", b), Matches, OK)
}

func (s *CmdSuite) TestFrontendCRUD(c *C) {
	b := "bk1"
	c.Assert(s.run("backend", "upsert", "-id", b), Matches, OK)

	f := "fr1"
	route := `Path("/path")`
	c.Assert(s.run(
		"frontend", "upsert",
		"-id", f, "-b", b, "-route", route,
		// Limits
		"-maxMemBodyKB", "6", "-maxBodyKB", "7",
		// Misc parameters
		// Failover predicate
		"-failoverPredicate", "IsNetworkError()",
		// Forward header
		"-trustForwardHeader",
		// Forward host
		"-forwardHost", "host1",
	),
		Matches, OK)

	fr, err := s.ng.GetFrontend(engine.FrontendKey{Id: f})
	c.Assert(err, IsNil)

	settings := fr.HTTPSettings()

	c.Assert(settings.Limits.MaxMemBodyBytes, Equals, int64(6*1024))
	c.Assert(settings.Limits.MaxBodyBytes, Equals, int64(7*1024))

	c.Assert(settings.FailoverPredicate, Equals, "IsNetworkError()")
	c.Assert(settings.TrustForwardHeader, Equals, true)
	c.Assert(settings.Hostname, Equals, "host1")

	c.Assert(s.run("frontend", "ls"), Matches, fmt.Sprintf(".*%v.*", f))
	c.Assert(s.run("frontend", "show", "-id", f), Matches, fmt.Sprintf(".*%v.*", f))
	c.Assert(s.run("frontend", "rm", "-id", f), Matches, OK)
}

func (s *CmdSuite) TestLimitsCRUD(c *C) {
	b := "bk1"
	c.Assert(s.run("backend", "upsert", "-id", b), Matches, OK)

	f := "fr1"
	route := `Path("/path")`
	c.Assert(s.run("frontend", "upsert", "-id", f, "-b", b, "-route", route), Matches, OK)

	rl := "rl1"
	c.Assert(s.run("ratelimit", "upsert", "-f", f, "-id", rl, "-requests", "10", "-variable", "client.ip", "-period", "3"), Matches, OK)
	c.Assert(s.run("ratelimit", "upsert", "-f", f, "-id", rl, "-requests", "100", "-variable", "client.ip", "-period", "30"), Matches, OK)
	c.Assert(s.run("ratelimit", "rm", "-f", f, "-id", rl), Matches, OK)

	cl := "cl1"
	c.Assert(s.run("connlimit", "upsert", "-f", f, "-id", cl, "-connections", "10", "-variable", "client.ip"), Matches, OK)
	c.Assert(s.run("connlimit", "upsert", "-f", f, "-id", cl, "-connections", "100", "-variable", "client.ip"), Matches, OK)

	fk := engine.FrontendKey{Id: f}
	out, err := s.ng.GetMiddleware(engine.MiddlewareKey{Id: cl, FrontendKey: fk})
	c.Assert(err, IsNil)
	c.Assert(out.Id, Equals, cl)

	c.Assert(s.run("connlimit", "rm", "-f", f, "-id", cl), Matches, OK)
}

func (s *CmdSuite) TestReadKeyPair(c *C) {
	keyPair := testutils.NewTestKeyPair()

	key, err := secret.NewKeyString()
	c.Assert(err, IsNil)

	fKey, err := ioutil.TempFile("", "vulcand")
	c.Assert(err, IsNil)
	defer fKey.Close()
	fKey.Write(keyPair.Key)

	fCert, err := ioutil.TempFile("", "vulcand")
	c.Assert(err, IsNil)
	defer fCert.Close()
	fCert.Write(keyPair.Cert)

	fSealed, err := ioutil.TempFile("", "vulcand")
	c.Assert(err, IsNil)
	fSealed.Close()

	s.run("secret", "seal_keypair", "-privateKey", fKey.Name(), "-cert", fCert.Name(), "-sealKey", key, "-f", fSealed.Name())

	bytes, err := ioutil.ReadFile(fSealed.Name())
	c.Assert(err, IsNil)

	box, err := secret.NewBoxFromKeyString(key)
	c.Assert(err, IsNil)

	sealed, err := secret.SealedValueFromJSON(bytes)
	data, err := box.Open(sealed)
	c.Assert(err, IsNil)

	outKeyPair, err := engine.KeyPairFromJSON(data)
	c.Assert(err, IsNil)

	c.Assert(outKeyPair, DeepEquals, keyPair)
}

func (s *CmdSuite) TestNewKey(c *C) {
	fKey, err := ioutil.TempFile("", "vulcand")
	c.Assert(err, IsNil)
	fKey.Close()

	s.run("secret", "new_key", "-f", fKey.Name())

	bytes, err := ioutil.ReadFile(fKey.Name())
	c.Assert(err, IsNil)

	_, err = secret.NewBoxFromKeyString(string(bytes))
	c.Assert(err, IsNil)
}
