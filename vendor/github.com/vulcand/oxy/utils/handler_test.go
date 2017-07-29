package utils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"

	. "gopkg.in/check.v1"
)

type UtilsSuite struct{}

var _ = Suite(&UtilsSuite{})

func (s *UtilsSuite) TestDefaultHandlerErrors(c *C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.(http.Hijacker)
		conn, _, _ := h.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	request, err := http.NewRequest("GET", srv.URL, strings.NewReader(""))
	c.Assert(err, IsNil)

	_, err = http.DefaultTransport.RoundTrip(request)

	w := NewBufferWriter(NopWriteCloser(&bytes.Buffer{}))

	DefaultHandler.ServeHTTP(w, nil, err)

	c.Assert(w.Code, Equals, http.StatusBadGateway)
}
