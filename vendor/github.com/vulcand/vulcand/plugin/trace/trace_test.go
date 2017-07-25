package trace

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/codegangsta/cli"
	"github.com/vulcand/oxy/testutils"
	oxytrace "github.com/vulcand/oxy/trace"
	"github.com/vulcand/vulcand/plugin"
	. "gopkg.in/check.v1"
)

func TestTrace(t *testing.T) { TestingT(t) }

type TraceSuite struct {
}

var _ = Suite(&TraceSuite{})

// One of the most important tests:
// Make sure the Rewrite spec is compatible and will be accepted by middleware registry
func (s *TraceSuite) TestSpecIsOK(c *C) {
	c.Assert(plugin.NewRegistry().AddSpec(GetSpec()), IsNil)
}

func (s *TraceSuite) TestGoodAddr(c *C) {
	vals := []string{
		// host + port format
		"syslog://localhost:5000",
		"syslog://localhost:5000?f=MAIL&sev=INFO",
		"syslog://localhost:5000?f=MAIL",
		"syslog://localhost:5000?f=LOG_LOCAL0&sev=DEBUG",

		// local socket format
		"syslog:///dev/log",
		"syslog:///dev/log?f=MAIL",
		"syslog:///dev/log?f=LOG_LOCAL0",

		// default syslog
		"syslog://",
		"syslog://?f=LOG_LOCAL0&sev=INFO",
	}
	for _, v := range vals {
		out, err := newWriter(v)
		c.Assert(err, IsNil)
		c.Assert(out, NotNil)
	}
}

func (s *TraceSuite) TestBadAddr(c *C) {
	vals := []string{
		"omglog://",
		"syslog://localhost:5000?f=SHMAIL",
		"syslog://localhost:5000?sev=SHMEVERITY",
	}
	for _, v := range vals {
		out, err := newWriter(v)
		c.Assert(err, NotNil)
		c.Assert(out, IsNil)
	}
}

func (s *TraceSuite) TestHandler(c *C) {
	os.Remove("/tmp/vulcand_trace_test.sock")
	unixAddr, err := net.ResolveUnixAddr("unixgram", "/tmp/vulcand_trace_test.sock")
	c.Assert(err, IsNil)
	conn, err := net.ListenUnixgram("unixgram", unixAddr)
	c.Assert(err, IsNil)
	defer conn.Close()

	outC := make(chan []byte, 1000)
	closeC := make(chan bool)
	defer close(closeC)
	go func() {
		for {
			buf := make([]byte, 65536)
			bytes, err := conn.Read(buf)
			if err != nil {
				return
			}
			outbuf := make([]byte, bytes)
			copy(outbuf, buf)
			select {
			case <-closeC:
				return
			case outC <- outbuf:
				continue
			}
		}
	}()

	responder := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("X-Resp-A", "h2")
		w.Write([]byte("hello"))
	})

	h, err := New("syslog:///tmp/vulcand_trace_test.sock", []string{"X-Req-A"}, []string{"X-Resp-A"})
	c.Assert(err, IsNil)

	handler, err := h.NewHandler(responder)
	c.Assert(err, IsNil)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL+"/hello", testutils.Header("X-Req-A", "yo"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	var buf []byte
	select {
	case buf = <-outC:
	case <-time.After(10 * time.Millisecond):
		c.Fatalf("timeout")
	}

	vals := strings.Split(string(buf), SyslogPrefix)
	var r *oxytrace.Record
	c.Assert(json.Unmarshal([]byte(vals[1]), &r), IsNil)
	c.Assert(r.Request.URL, Equals, "/hello")
	c.Assert(r.Request.Headers, DeepEquals, http.Header{"X-Req-A": []string{"yo"}})
	c.Assert(r.Response.Headers, DeepEquals, http.Header{"X-Resp-A": []string{"h2"}})
}

func (s *TraceSuite) TestNewFromCLI(c *C) {
	app := cli.NewApp()
	app.Name = "test"
	executed := false
	app.Action = func(ctx *cli.Context) {
		executed = true
		out, err := FromCli(ctx)
		c.Assert(out, NotNil)
		c.Assert(err, IsNil)

		t := out.(*Trace)
		c.Assert(t.Addr, Equals, "syslog:///dev/log?sev=INFO&f=MAIL")
		c.Assert(t.ReqHeaders, DeepEquals, []string{"X-A", "X-B"})
		c.Assert(t.RespHeaders, DeepEquals, []string{"X-C", "X-D"})
	}
	app.Flags = CliFlags()
	app.Run([]string{"test", "--addr=syslog:///dev/log?sev=INFO&f=MAIL", "--reqHeader=X-A", "--reqHeader=X-B", "--respHeader=X-C", "--respHeader=X-D"})
	c.Assert(executed, Equals, true)
}
