package rewrite

import (
	"bytes"
	"net/http"
	"strings"

	. "gopkg.in/check.v1"
)

type TemplateSuite struct{}

var _ = Suite(&TemplateSuite{})

func (s *TemplateSuite) SetUpSuite(c *C) {
}

func (s *TemplateSuite) TestTemplateOkay(c *C) {
	request, _ := http.NewRequest("GET", "http://foo", nil)
	request.Header.Add("X-Header", "bar")

	out := &bytes.Buffer{}
	err := Apply(strings.NewReader(`foo {{.Request.Header.Get "X-Header"}}`), out, request)
	c.Assert(err, IsNil)
	c.Assert(out.String(), Equals, "foo bar")
}

func (s *TemplateSuite) TestBadTemplate(c *C) {
	request, _ := http.NewRequest("GET", "http://foo", nil)
	request.Header.Add("X-Header", "bar")

	out := &bytes.Buffer{}
	err := Apply(strings.NewReader(`foo {{.Request.Header.Get "X-Header"`), out, request)
	c.Assert(err, NotNil)
	c.Assert(out.String(), Equals, "")
}

func (s *TemplateSuite) TestNoVariables(c *C) {
	request, _ := http.NewRequest("GET", "http://foo", nil)
	request.Header.Add("X-Header", "bar")

	out := &bytes.Buffer{}
	err := Apply(strings.NewReader(`foo baz`), out, request)
	c.Assert(err, IsNil)
	c.Assert(out.String(), Equals, "foo baz")
}

func (s *TemplateSuite) TestNonexistentVariable(c *C) {
	request, _ := http.NewRequest("GET", "http://foo", nil)
	request.Header.Add("X-Header", "bar")

	out := &bytes.Buffer{}
	err := Apply(strings.NewReader(`foo {{.Request.Header.Get "Y-Header"}}`), out, request)
	c.Assert(err, IsNil)
	c.Assert(out.String(), Equals, "foo ")
}
