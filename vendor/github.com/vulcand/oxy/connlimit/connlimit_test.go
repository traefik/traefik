package connlimit

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vulcand/oxy/testutils"
	"github.com/vulcand/oxy/utils"

	. "gopkg.in/check.v1"
)

func TestConn(t *testing.T) { TestingT(t) }

type ConnLimiterSuite struct {
}

var _ = Suite(&ConnLimiterSuite{})

func (s *ConnLimiterSuite) SetUpSuite(c *C) {
}

// We've hit the limit and were able to proceed once the request has completed
func (s *ConnLimiterSuite) TestHitLimitAndRelease(c *C) {
	wait := make(chan bool)
	proceed := make(chan bool)
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("wait") != "" {
			proceed <- true
			<-wait
		}
		w.Write([]byte("hello"))
	})

	l, err := New(handler, headerLimit, 1)
	c.Assert(err, Equals, nil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	go func() {
		re, _, err := testutils.Get(srv.URL, testutils.Header("Limit", "a"), testutils.Header("wait", "yes"))
		c.Assert(err, IsNil)
		c.Assert(re.StatusCode, Equals, http.StatusOK)
	}()

	<-proceed

	re, _, err := testutils.Get(srv.URL, testutils.Header("Limit", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	// request from another source succeeds
	re, _, err = testutils.Get(srv.URL, testutils.Header("Limit", "b"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	// Once the first request finished, next one succeeds
	close(wait)

	re, _, err = testutils.Get(srv.URL, testutils.Header("Limit", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// We've hit the limit and were able to proceed once the request has completed
func (s *ConnLimiterSuite) TestCustomHandlers(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	errHandler := utils.ErrorHandlerFunc(func(w http.ResponseWriter, req *http.Request, err error) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(http.StatusText(http.StatusTeapot)))
	})

	buf := &bytes.Buffer{}
	log := utils.NewFileLogger(buf, utils.INFO)

	l, err := New(handler, headerLimit, 0, ErrorHandler(errHandler), Logger(log))
	c.Assert(err, Equals, nil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL, testutils.Header("Limit", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusTeapot)

	c.Assert(len(buf.String()), Not(Equals), 0)
}

// We've hit the limit and were able to proceed once the request has completed
func (s *ConnLimiterSuite) TestFaultyExtract(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	l, err := New(handler, faultyExtract, 1)
	c.Assert(err, Equals, nil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusInternalServerError)
}

func headerLimiter(req *http.Request) (string, int64, error) {
	return req.Header.Get("Limit"), 1, nil
}

func faultyExtractor(req *http.Request) (string, int64, error) {
	return "", -1, fmt.Errorf("oops")
}

var headerLimit = utils.ExtractorFunc(headerLimiter)
var faultyExtract = utils.ExtractorFunc(faultyExtractor)
