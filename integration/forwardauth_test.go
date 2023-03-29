package integration

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v3/integration/try"
	checker "github.com/vdemeester/shakers"
)

const (
	portAuth                   = "9001"
	portBackend                = "9000"
	cookieNameAuth             = "Cookie-Auth"
	cookieNameAuthNotForwarded = "Cookie-Auth-Not-Forwarded"
	cookieNameBoth             = "Cookie-Both"
	cookieNameBackend          = "Cookie-Backend"
	cookieValueAuth            = "Auth"
	cookieValueBackend         = "Backend"
	headerBackend              = "Foo"
	headerValueBackend         = "baz"
	userID                     = "123"
)

// ForwardAuth tests suite.
type ForwardAuthSuite struct{ BaseSuite }

func (s *ForwardAuthSuite) TestBasic(c *check.C) {
	params := struct{ PortAuth, PortBackend string }{portAuth, portBackend}
	file := s.adaptFile(c, "fixtures/forwardauth/basic.toml", params)
	defer os.Remove(file)
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	auth := startTestServerWithHandler(newAuthHandler(), portAuth)
	defer auth.Close()

	backend := startTestServerWithHandler(newBackendHandler(), portBackend)
	defer backend.Close()

	expectedHeaders := map[string][]string{
		headerBackend: {headerValueBackend},
		"Set-Cookie": {
			cookieNameBackend + "=" + cookieValueBackend,
			cookieNameAuth + "=" + cookieValueAuth,
			cookieNameBoth + "=" + cookieValueAuth,
		},
	}

	err = try.GetRequest(
		"http://127.0.0.1:8000/backend",
		1000*time.Millisecond,
		try.StatusCodeIs(http.StatusOK),
		try.HasHeaderStruct(expectedHeaders),
		try.BodyContains(userID),
	)
	c.Assert(err, checker.IsNil)
}

func startTestServerWithHandler(h http.Handler, port string) (ts *httptest.Server) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		panic(err)
	}
	ts = &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: h},
	}
	ts.Start()
	return ts
}

func newAuthHandler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: cookieNameAuth, Value: cookieValueAuth})
		http.SetCookie(w, &http.Cookie{Name: cookieNameAuthNotForwarded, Value: cookieValueAuth})
		http.SetCookie(w, &http.Cookie{Name: cookieNameBoth, Value: cookieValueAuth})
		w.WriteHeader(http.StatusOK)
	})
	return m
}

func newBackendHandler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/backend", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: cookieNameBoth, Value: cookieValueBackend})
		http.SetCookie(w, &http.Cookie{Name: cookieNameBackend, Value: cookieValueBackend})
		w.Header().Set(headerBackend, headerValueBackend)
		fmt.Fprint(w, userID)
	})
	return m
}
