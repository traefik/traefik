package forward

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mailgun/oxy/testutils"
	"github.com/mailgun/oxy/utils"

	. "gopkg.in/check.v1"
)

func TestFwd(t *testing.T) { TestingT(t) }

type FwdSuite struct{}

var _ = Suite(&FwdSuite{})

// Makes sure hop-by-hop headers are removed
func (s *FwdSuite) TestForwardHopHeaders(c *C) {
	called := false
	var outHeaders http.Header
	var outHost, expectedHost string
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		called = true
		outHeaders = req.Header
		outHost = req.Host
		w.Write([]byte("hello"))
	})
	defer srv.Close()

	f, err := New()
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		expectedHost = req.URL.Host
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	headers := http.Header{
		Connection: []string{"close"},
		KeepAlive:  []string{"timeout=600"},
	}

	re, body, err := testutils.Get(proxy.URL, testutils.Headers(headers))
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "hello")
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(called, Equals, true)
	c.Assert(outHeaders.Get(Connection), Equals, "")
	c.Assert(outHeaders.Get(KeepAlive), Equals, "")
	c.Assert(outHost, Equals, expectedHost)
}

func (s *FwdSuite) TestDefaultErrHandler(c *C) {
	f, err := New()
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI("http://localhost:63450")
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	re, _, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusBadGateway)
}

func (s *FwdSuite) TestCustomErrHandler(c *C) {
	f, err := New(ErrorHandler(utils.ErrorHandlerFunc(func(w http.ResponseWriter, req *http.Request, err error) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(http.StatusText(http.StatusTeapot)))
	})))
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI("http://localhost:63450")
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	re, body, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusTeapot)
	c.Assert(string(body), Equals, http.StatusText(http.StatusTeapot))
}

// Makes sure hop-by-hop headers are removed
func (s *FwdSuite) TestForwardedHeaders(c *C) {
	var outHeaders http.Header
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		outHeaders = req.Header
		w.Write([]byte("hello"))
	})
	defer srv.Close()

	f, err := New(Rewriter(&HeaderRewriter{TrustForwardHeader: true, Hostname: "hello"}))
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	headers := http.Header{
		XForwardedProto:  []string{"httpx"},
		XForwardedFor:    []string{"192.168.1.1"},
		XForwardedServer: []string{"foobar"},
		XForwardedHost:   []string{"upstream-foobar"},
	}

	re, _, err := testutils.Get(proxy.URL, testutils.Headers(headers))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(outHeaders.Get(XForwardedProto), Equals, "httpx")
	c.Assert(strings.Contains(outHeaders.Get(XForwardedFor), "192.168.1.1"), Equals, true)
	c.Assert(strings.Contains(outHeaders.Get(XForwardedHost), "upstream-foobar"), Equals, true)
	c.Assert(outHeaders.Get(XForwardedServer), Equals, "hello")
}

func (s *FwdSuite) TestCustomRewriter(c *C) {
	var outHeaders http.Header
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		outHeaders = req.Header
		w.Write([]byte("hello"))
	})
	defer srv.Close()

	f, err := New(Rewriter(&HeaderRewriter{TrustForwardHeader: false, Hostname: "hello"}))
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	headers := http.Header{
		XForwardedProto: []string{"httpx"},
		XForwardedFor:   []string{"192.168.1.1"},
	}

	re, _, err := testutils.Get(proxy.URL, testutils.Headers(headers))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(outHeaders.Get(XForwardedProto), Equals, "http")
	c.Assert(strings.Contains(outHeaders.Get(XForwardedFor), "192.168.1.1"), Equals, false)
}

func (s *FwdSuite) TestCustomTransportTimeout(c *C) {
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.Write([]byte("hello"))
	})
	defer srv.Close()

	f, err := New(RoundTripper(
		&http.Transport{
			ResponseHeaderTimeout: 5 * time.Millisecond,
		}))
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	re, _, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusGatewayTimeout)
}

func (s *FwdSuite) TestCustomLogger(c *C) {
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})
	defer srv.Close()

	buf := &bytes.Buffer{}
	l := utils.NewFileLogger(buf, utils.INFO)

	f, err := New(Logger(l))
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	re, _, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(strings.Contains(buf.String(), srv.URL), Equals, true)
}

func (s *FwdSuite) TestEscapedURL(c *C) {
	var outURL string
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		outURL = req.RequestURI
		w.Write([]byte("hello"))
	})
	defer srv.Close()

	f, err := New()
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	path := "/log/http%3A%2F%2Fwww.site.com%2Fsomething?a=b"

	request, err := http.NewRequest("GET", proxy.URL, nil)
	parsed := testutils.ParseURI(proxy.URL)
	parsed.Opaque = path
	request.URL = parsed
	re, err := http.DefaultClient.Do(request)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(outURL, Equals, path)
}

func (s *FwdSuite) TestForwardedProto(c *C) {
	var proto string
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		proto = req.Header.Get(XForwardedProto)
		w.Write([]byte("hello"))
	})
	defer srv.Close()

	buf := &bytes.Buffer{}
	l := utils.NewFileLogger(buf, utils.INFO)

	f, err := New(Logger(l))
	c.Assert(err, IsNil)

	proxy := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		f.ServeHTTP(w, req)
	})
	tproxy := httptest.NewUnstartedServer(proxy)
	tproxy.StartTLS()
	defer tproxy.Close()

	re, _, err := testutils.Get(tproxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(proto, Equals, "https")

	c.Assert(strings.Contains(buf.String(), "tls"), Equals, true)
}

func (s *FwdSuite) TestChunkedResponseConversion(c *C) {
	srv := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		h := w.(http.Hijacker)
		conn, _, _ := h.Hijack()
		fmt.Fprintf(conn, "HTTP/1.0 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n4\r\ntest\r\n5\r\ntest1\r\n5\r\ntest2\r\n0\r\n\r\n")
		conn.Close()
	})
	defer srv.Close()

	f, err := New()
	c.Assert(err, IsNil)

	proxy := testutils.NewHandler(func(w http.ResponseWriter, req *http.Request) {
		req.URL = testutils.ParseURI(srv.URL)
		f.ServeHTTP(w, req)
	})
	defer proxy.Close()

	re, body, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "testtest1test2")
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(re.Header.Get("Content-Length"), Equals, fmt.Sprintf("%d", len("testtest1test2")))
}
