package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

// ErrorPagesSuite test suites.
type ErrorPagesSuite struct {
	BaseSuite
	ErrorPageIP string
	BackendIP   string
}

func (s *ErrorPagesSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "error_pages")
	s.composeUp(c)

	s.ErrorPageIP = s.getComposeServiceIP(c, "nginx2")
	s.BackendIP = s.getComposeServiceIP(c, "nginx1")
}

func (s *ErrorPagesSuite) TestSimpleConfiguration(c *check.C) {
	file := s.adaptFile(c, "fixtures/error_pages/simple.toml", struct {
		Server1 string
		Server2 string
	}{"http://" + s.BackendIP + ":80", s.ErrorPageIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.local"

	err = try.Request(frontendReq, 2*time.Second, try.BodyContains("nginx"))
	c.Assert(err, checker.IsNil)
}

func (s *ErrorPagesSuite) TestErrorPage(c *check.C) {
	// error.toml contains a mis-configuration of the backend host
	file := s.adaptFile(c, "fixtures/error_pages/error.toml", struct {
		Server1 string
		Server2 string
	}{s.BackendIP, s.ErrorPageIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.local"

	err = try.Request(frontendReq, 2*time.Second, try.BodyContains("An error occurred."))
	c.Assert(err, checker.IsNil)
}

func (s *ErrorPagesSuite) TestErrorPageFlush(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Transfer-Encoding", "chunked")
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte("KO"))
	}))

	file := s.adaptFile(c, "fixtures/error_pages/simple.toml", struct {
		Server1 string
		Server2 string
	}{srv.URL, s.ErrorPageIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.local"

	err = try.Request(frontendReq, 2*time.Second,
		try.BodyContains("An error occurred."),
		try.HasHeaderValue("Content-Type", "text/html", true),
	)
	c.Assert(err, checker.IsNil)
}
