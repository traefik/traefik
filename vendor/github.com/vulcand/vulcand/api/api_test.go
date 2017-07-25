package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/scroll"
	oxytest "github.com/vulcand/oxy/testutils"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/engine/memng"
	"github.com/vulcand/vulcand/plugin/connlimit"
	"github.com/vulcand/vulcand/plugin/registry"
	"github.com/vulcand/vulcand/proxy"
	"github.com/vulcand/vulcand/stapler"
	"github.com/vulcand/vulcand/supervisor"
	"github.com/vulcand/vulcand/testutils"

	. "gopkg.in/check.v1"
)

func TestApi(t *testing.T) { TestingT(t) }

type ApiSuite struct {
	ng         engine.Engine
	testServer *httptest.Server
	client     *Client
}

var _ = Suite(&ApiSuite{})

func (s *ApiSuite) SetUpTest(c *C) {
	newProxy := func(id int) (proxy.Proxy, error) {
		return proxy.New(id, stapler.New(), proxy.Options{})
	}

	s.ng = memng.New(registry.GetRegistry())

	sv := supervisor.New(newProxy, s.ng, make(chan error), supervisor.Options{})

	app := scroll.NewApp()
	InitProxyController(s.ng, sv, app)
	s.testServer = httptest.NewServer(app.GetHandler())
	s.client = NewClient(s.testServer.URL, registry.GetRegistry())
}

func (s *ApiSuite) TearDownTest(c *C) {
	s.testServer.Close()
}

func (s *ApiSuite) TestStatus(c *C) {
	c.Assert(s.client.GetStatus(), IsNil)
}

func (s *ApiSuite) TestStatusV1(c *C) {
	re, body, err := oxytest.Get(s.testServer.URL + "/v1/status")
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(string(body), Equals, `{"Status":"ok"}`)
}

func (s *ApiSuite) TestSeverity(c *C) {
	for _, sev := range []log.Level{log.InfoLevel, log.WarnLevel, log.ErrorLevel} {
		err := s.client.UpdateLogSeverity(sev)
		c.Assert(err, IsNil)
		out, err := s.client.GetLogSeverity()
		c.Assert(err, IsNil)
		c.Assert(out, Equals, sev)
	}
}

func (s *ApiSuite) TestInvalidSeverity(c *C) {
	err := s.client.UpdateLogSeverity(255)
	c.Assert(err, NotNil)
}

func (s *ApiSuite) TestNotFoundHandler(c *C) {
	_, err := s.client.Get(s.client.endpoint("blabla"), nil)
	c.Assert(err, FitsTypeOf, &engine.NotFoundError{})
}

func (s *ApiSuite) TestHostCRUD(c *C) {
	host := engine.Host{Name: "localhost"}
	c.Assert(s.client.UpsertHost(host), IsNil)

	hosts, _ := s.ng.GetHosts()
	c.Assert(len(hosts), Equals, 1)

	hosts, err := s.client.GetHosts()
	c.Assert(hosts, NotNil)
	c.Assert(err, IsNil)
	c.Assert(hosts[0].Name, Equals, "localhost")

	out, err := s.client.GetHost(engine.HostKey{Name: host.Name})
	c.Assert(err, IsNil)
	c.Assert(out.Name, Equals, host.Name)

	host.Settings.KeyPair = testutils.NewTestKeyPair()
	c.Assert(s.client.UpsertHost(host), IsNil)

	out, err = s.ng.GetHost(engine.HostKey{Name: host.Name})
	c.Assert(out.Settings.KeyPair, DeepEquals, host.Settings.KeyPair)

	err = s.client.DeleteHost(engine.HostKey{Name: host.Name})
	c.Assert(err, IsNil)

	hosts, _ = s.ng.GetHosts()
	c.Assert(len(hosts), Equals, 0)

	hosts, err = s.client.GetHosts()
	c.Assert(len(hosts), Equals, 0)
	c.Assert(err, IsNil)
}

func (s *ApiSuite) TestHostDeleteBad(c *C) {
	err := s.client.DeleteHost(engine.HostKey{Name: "localhost"})
	c.Assert(err, FitsTypeOf, &engine.NotFoundError{})
}

func (s *ApiSuite) TestBackendCRUD(c *C) {
	b, err := engine.NewHTTPBackend("b1", engine.HTTPBackendSettings{})
	c.Assert(err, IsNil)

	c.Assert(s.client.UpsertBackend(*b), IsNil)

	bs, _ := s.ng.GetBackends()
	c.Assert(len(bs), Equals, 1)
	c.Assert(bs[0], DeepEquals, *b)

	bs, err = s.client.GetBackends()
	c.Assert(bs, NotNil)
	c.Assert(err, IsNil)
	c.Assert(bs[0], DeepEquals, *b)

	bk := engine.BackendKey{Id: b.Id}
	out, err := s.client.GetBackend(bk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, b)

	settings := b.HTTPSettings()
	settings.Timeouts.Read = "1s"
	b.Settings = settings
	c.Assert(s.client.UpsertBackend(*b), IsNil)

	out, err = s.client.GetBackend(bk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, b)

	err = s.client.DeleteBackend(bk)
	c.Assert(err, IsNil)

	out, err = s.client.GetBackend(bk)
	c.Assert(err, FitsTypeOf, &engine.NotFoundError{})
}

func (s *ApiSuite) TestServerCRUD(c *C) {
	b, err := engine.NewHTTPBackend("b1", engine.HTTPBackendSettings{})
	c.Assert(err, IsNil)

	c.Assert(s.client.UpsertBackend(*b), IsNil)

	srv1 := engine.Server{Id: "srv1", URL: "http://localhost:5000"}
	srv2 := engine.Server{Id: "srv2", URL: "http://localhost:6000"}

	bk := engine.BackendKey{Id: b.Id}
	c.Assert(s.client.UpsertServer(bk, srv1, 0), IsNil)
	c.Assert(s.client.UpsertServer(bk, srv2, 0), IsNil)

	srvs, _ := s.ng.GetServers(bk)
	c.Assert(len(srvs), Equals, 2)
	c.Assert(srvs[0], DeepEquals, srv1)
	c.Assert(srvs[1], DeepEquals, srv2)

	srvs, err = s.client.GetServers(bk)
	c.Assert(srvs, NotNil)
	c.Assert(len(srvs), Equals, 2)
	c.Assert(srvs[0], DeepEquals, srv1)
	c.Assert(srvs[1], DeepEquals, srv2)

	sk := engine.ServerKey{Id: srv1.Id, BackendKey: bk}
	out, err := s.client.GetServer(sk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &srv1)

	srv1.URL = "http://localhost:5001"
	c.Assert(s.client.UpsertServer(bk, srv1, 0), IsNil)

	out, err = s.client.GetServer(sk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &srv1)

	err = s.client.DeleteServer(sk)
	c.Assert(err, IsNil)

	out, err = s.client.GetServer(sk)
	c.Assert(err, FitsTypeOf, &engine.NotFoundError{})
}

func (s *ApiSuite) TestFrontendCRUD(c *C) {
	b, err := engine.NewHTTPBackend("b1", engine.HTTPBackendSettings{})
	c.Assert(err, IsNil)

	c.Assert(s.client.UpsertBackend(*b), IsNil)

	f, err := engine.NewHTTPFrontend(s.ng.GetRegistry().GetRouter(), "f1", b.Id, `Path("/")`, engine.HTTPFrontendSettings{})
	c.Assert(err, IsNil)
	fk := engine.FrontendKey{Id: f.Id}

	c.Assert(s.client.UpsertFrontend(*f, 0), IsNil)

	fs, err := s.client.GetFrontends()
	c.Assert(err, IsNil)
	c.Assert(fs[0], DeepEquals, *f)

	out, err := s.client.GetFrontend(fk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, f)

	settings := f.HTTPSettings()
	settings.Hostname = `localhost`
	f.Settings = settings
	f.Route = `Path("/v2")`

	c.Assert(s.client.UpsertFrontend(*f, 0), IsNil)

	c.Assert(s.client.DeleteFrontend(fk), IsNil)

	out, err = s.client.GetFrontend(fk)
	c.Assert(err, FitsTypeOf, &engine.NotFoundError{})
}

func (s *ApiSuite) TestListenerCRUD(c *C) {
	l := engine.Listener{Id: "l1", Address: engine.Address{Network: "tcp", Address: "localhost:1300"}, Protocol: engine.HTTP}

	c.Assert(s.client.UpsertListener(l), IsNil)

	ls, err := s.client.GetListeners()
	c.Assert(err, IsNil)
	c.Assert(ls[0], DeepEquals, l)

	lk := engine.ListenerKey{Id: l.Id}
	out, err := s.client.GetListener(lk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &l)

	c.Assert(s.client.DeleteListener(lk), IsNil)

	out, err = s.client.GetListener(lk)
	c.Assert(err, FitsTypeOf, &engine.NotFoundError{})
}

func (s *ApiSuite) TestMiddlewareCRUD(c *C) {
	b, err := engine.NewHTTPBackend("b1", engine.HTTPBackendSettings{})
	c.Assert(err, IsNil)

	c.Assert(s.client.UpsertBackend(*b), IsNil)

	f, err := engine.NewHTTPFrontend(s.ng.GetRegistry().GetRouter(), "f1", b.Id, `Path("/")`, engine.HTTPFrontendSettings{})
	c.Assert(err, IsNil)
	fk := engine.FrontendKey{Id: f.Id}

	c.Assert(s.client.UpsertFrontend(*f, 0), IsNil)

	cl := s.makeConnLimit("c1", 10, "client.ip", 2, f)
	c.Assert(s.client.UpsertMiddleware(fk, cl, 0), IsNil)

	ms, err := s.client.GetMiddlewares(fk)
	c.Assert(err, IsNil)
	c.Assert(ms[0], DeepEquals, cl)

	cl = s.makeConnLimit("c1", 10, "client.ip", 3, f)
	c.Assert(s.client.UpsertMiddleware(fk, cl, 0), IsNil)

	mk := engine.MiddlewareKey{Id: cl.Id, FrontendKey: fk}
	v, err := s.client.GetMiddleware(mk)
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &cl)

	c.Assert(s.client.DeleteMiddleware(mk), IsNil)

	_, err = s.client.GetMiddleware(mk)
	c.Assert(err, FitsTypeOf, &engine.NotFoundError{})

}

func (s *ApiSuite) makeConnLimit(id string, connections int64, variable string, priority int, f *engine.Frontend) engine.Middleware {
	cl, err := connlimit.NewConnLimit(connections, variable)
	if err != nil {
		panic(err)
	}
	return engine.Middleware{
		Type:       "connlimit",
		Id:         id,
		Middleware: cl,
	}
}
