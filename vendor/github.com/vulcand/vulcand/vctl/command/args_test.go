package command

import (
	. "gopkg.in/check.v1"
)

type ArgsSuite struct {
}

var _ = Suite(&ArgsSuite{})

func (s *ArgsSuite) TestFindVulcanUrl(c *C) {
	url, args, err := findVulcanUrl([]string{"vctl", "--vulcan=bla"})
	c.Assert(err, IsNil)
	c.Assert(url, Equals, "bla")
	c.Assert(args, DeepEquals, []string{"vctl"})
}

func (s *ArgsSuite) TestFindDefaults(c *C) {
	url, args, err := findVulcanUrl([]string{"vctl", "status"})
	c.Assert(err, IsNil)
	c.Assert(url, Equals, "http://localhost:8182")
	c.Assert(args, DeepEquals, []string{"vctl", "status"})
}

func (s *ArgsSuite) TestFindMiddle(c *C) {
	url, args, err := findVulcanUrl([]string{"vctl", "endpoint", "-vulcan", "http://yo", "rm"})
	c.Assert(err, IsNil)
	c.Assert(url, Equals, "http://yo")
	c.Assert(args, DeepEquals, []string{"vctl", "endpoint", "rm"})
}

func (s *ArgsSuite) TestFindNoUrl(c *C) {
	_, _, err := findVulcanUrl([]string{"vctl", "endpoint", "rm", "-vulcan"})
	c.Assert(err, NotNil)
}
