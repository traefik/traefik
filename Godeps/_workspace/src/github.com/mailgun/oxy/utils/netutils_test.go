package utils

import (
	"net/http"
	"net/url"
	"testing"

	. "gopkg.in/check.v1"
)

func TestUtils(t *testing.T) { TestingT(t) }

type NetUtilsSuite struct{}

var _ = Suite(&NetUtilsSuite{})

// Make sure copy does it right, so the copied url
// is safe to alter without modifying the other
func (s *NetUtilsSuite) TestCopyUrl(c *C) {
	urlA := &url.URL{
		Scheme:   "http",
		Host:     "localhost:5000",
		Path:     "/upstream",
		Opaque:   "opaque",
		RawQuery: "a=1&b=2",
		Fragment: "#hello",
		User:     &url.Userinfo{},
	}
	urlB := CopyURL(urlA)
	c.Assert(urlB, DeepEquals, urlA)
	urlB.Scheme = "https"
	c.Assert(urlB, Not(DeepEquals), urlA)
}

// Make sure copy headers is not shallow and copies all headers
func (s *NetUtilsSuite) TestCopyHeaders(c *C) {
	source, destination := make(http.Header), make(http.Header)
	source.Add("a", "b")
	source.Add("c", "d")

	CopyHeaders(destination, source)

	c.Assert(destination.Get("a"), Equals, "b")
	c.Assert(destination.Get("c"), Equals, "d")

	// make sure that altering source does not affect the destination
	source.Del("a")
	c.Assert(source.Get("a"), Equals, "")
	c.Assert(destination.Get("a"), Equals, "b")
}

func (s *NetUtilsSuite) TestHasHeaders(c *C) {
	source := make(http.Header)
	source.Add("a", "b")
	source.Add("c", "d")
	c.Assert(HasHeaders([]string{"a", "f"}, source), Equals, true)
	c.Assert(HasHeaders([]string{"i", "j"}, source), Equals, false)
}

func (s *NetUtilsSuite) TestRemoveHeaders(c *C) {
	source := make(http.Header)
	source.Add("a", "b")
	source.Add("a", "m")
	source.Add("c", "d")
	RemoveHeaders(source, "a")
	c.Assert(source.Get("a"), Equals, "")
	c.Assert(source.Get("c"), Equals, "d")
}
