package route

import (
	"net/http"
	"testing"

	. "gopkg.in/check.v1"
)

func TestMatcher(t *testing.T) { TestingT(t) }

type MatcherSuite struct {
}

var _ = Suite(&MatcherSuite{})

func (s *MatcherSuite) TestHostnameCase(c *C) {
	var matcher1, matcher2 matcher
	var req *http.Request
	var err error

	req, err = http.NewRequest("GET", "http://example.com", nil)
	c.Assert(err, IsNil)

	matcher1, err = hostTrieMatcher("example.com")
	c.Assert(err, IsNil)
	matcher2, err = hostTrieMatcher("Example.Com")
	c.Assert(err, IsNil)

	c.Assert(matcher1.match(req), Not(IsNil))
	c.Assert(matcher2.match(req), Not(IsNil))

	matcher1, err = hostRegexpMatcher(`.*example.com`)
	c.Assert(err, IsNil)
	matcher2, err = hostRegexpMatcher(`.*Example.Com`)
	c.Assert(err, IsNil)

	c.Assert(matcher1.match(req), Not(IsNil))
	c.Assert(matcher2.match(req), Not(IsNil))
}
