package route

import (
	. "gopkg.in/check.v1"
)

type UtilsSuite struct{}

var _ = Suite(&UtilsSuite{})

// Make sure parseUrl is strict enough not to accept total garbage
func (s *UtilsSuite) TestRawPath(c *C) {
	vals := []struct {
		URL      string
		Expected string
	}{
		{"http://google.com/", "/"},
		{"http://google.com/a?q=b", "/a"},
		{"http://google.com/%2Fvalue/hello", "/%2Fvalue/hello"},
		{"/home", "/home"},
		{"/home?a=b", "/home"},
		{"/home%2F", "/home%2F"},
	}
	for _, v := range vals {
		out := rawPath(makeReq(req{url: v.URL}))
		c.Assert(out, Equals, v.Expected)
	}
}
