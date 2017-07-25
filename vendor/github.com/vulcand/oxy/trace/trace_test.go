package trace

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/vulcand/oxy/testutils"
	"github.com/vulcand/oxy/utils"

	. "gopkg.in/check.v1"
)

func TestTrace(t *testing.T) { TestingT(t) }

type TraceSuite struct{}

var _ = Suite(&TraceSuite{})

func (s *TraceSuite) TestTraceSimple(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Length", "5")
		w.Write([]byte("hello"))
	})
	buf := &bytes.Buffer{}
	l := utils.NewFileLogger(buf, utils.INFO)

	trace := &bytes.Buffer{}

	t, err := New(handler, trace, Logger(l))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(t)
	defer srv.Close()

	re, _, err := testutils.MakeRequest(srv.URL+"/hello", testutils.Method("POST"), testutils.Body("123456"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	var r *Record
	c.Assert(json.Unmarshal(trace.Bytes(), &r), IsNil)

	c.Assert(r.Request.Method, Equals, "POST")
	c.Assert(r.Request.URL, Equals, "/hello")
	c.Assert(r.Response.Code, Equals, http.StatusOK)
	c.Assert(r.Request.BodyBytes, Equals, int64(6))
	c.Assert(r.Response.Roundtrip, Not(Equals), float64(0))
	c.Assert(r.Response.BodyBytes, Equals, int64(5))
}

func (s *TraceSuite) TestTraceCaptureHeaders(c *C) {
	respHeaders := http.Header{
		"X-Re-1": []string{"6", "7"},
		"X-Re-2": []string{"2", "3"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		utils.CopyHeaders(w.Header(), respHeaders)
		w.Write([]byte("hello"))
	})
	buf := &bytes.Buffer{}
	l := utils.NewFileLogger(buf, utils.INFO)

	trace := &bytes.Buffer{}

	t, err := New(handler, trace, Logger(l), RequestHeaders("X-Req-B", "X-Req-A"), ResponseHeaders("X-Re-1", "X-Re-2"))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(t)
	defer srv.Close()

	reqHeaders := http.Header{"X-Req-A": []string{"1", "2"}, "X-Req-B": []string{"3", "4"}}
	re, _, err := testutils.Get(srv.URL+"/hello", testutils.Headers(reqHeaders))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	var r *Record
	c.Assert(json.Unmarshal(trace.Bytes(), &r), IsNil)

	c.Assert(r.Request.Headers, DeepEquals, reqHeaders)
	c.Assert(r.Response.Headers, DeepEquals, respHeaders)
}

func (s *TraceSuite) TestTraceTLS(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})
	buf := &bytes.Buffer{}
	l := utils.NewFileLogger(buf, utils.INFO)

	trace := &bytes.Buffer{}

	t, err := New(handler, trace, Logger(l))
	c.Assert(err, IsNil)

	srv := httptest.NewUnstartedServer(t)
	srv.StartTLS()
	defer srv.Close()

	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	u, err := url.Parse(srv.URL)
	c.Assert(err, IsNil)

	conn, err := tls.Dial("tcp", u.Host, config)
	c.Assert(err, IsNil)

	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	c.Assert(status, Equals, "HTTP/1.0 200 OK\r\n")
	state := conn.ConnectionState()
	conn.Close()

	var r *Record
	c.Assert(json.Unmarshal(trace.Bytes(), &r), IsNil)
	c.Assert(r.Request.TLS.Version, Equals, versionToString(state.Version))
}
