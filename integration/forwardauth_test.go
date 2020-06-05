package integration

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"

	"github.com/containous/traefik/v2/integration/try"
)

const portAuth = "9001"
const portBackend = "9000"
const cookieNameAuth = "Foo"
const cookieNameBoth = "Bar"
const headerBoth = "Baz"
const headerAuth = "X-User-Id"
const headerValueAuth = "Auth"
const headerValueBackend = "Backend"
const userID = "123"

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
	defer cmd.Process.Kill()

	auth := startTestServerWithHandler(newAuthHandler(), portAuth)
	defer auth.Close()

	backend := startTestServerWithHandler(newBackendHandler(), portBackend)
	defer backend.Close()

	expectedHeaders := map[string][]string{
		"Set-Cookie": {
			cookieNameAuth + "=" + headerValueAuth,
			cookieNameBoth + "=" + headerValueBackend,
		},
		headerBoth: {headerValueBackend},
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
		http.SetCookie(w, &http.Cookie{Name: cookieNameAuth, Value: headerValueAuth})
		http.SetCookie(w, &http.Cookie{Name: cookieNameBoth, Value: headerValueAuth})
		w.Header().Set(headerBoth, headerValueAuth)
		w.Header().Set(headerAuth, userID)
		w.WriteHeader(http.StatusOK)
	})
	return m
}

func newBackendHandler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/backend", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: cookieNameBoth, Value: headerValueBackend})
		w.Header().Set(headerBoth, headerValueBackend)
		fmt.Fprint(w, r.Header.Get(headerAuth))
	})
	return m
}
