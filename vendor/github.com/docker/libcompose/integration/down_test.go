package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestDown(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")

	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)
	c.Assert(cn.State.Running, Equals, true)

	s.FromText(c, p, "down", SimpleTemplate)

	containers := s.GetContainersByProject(c, p)
	c.Assert(len(containers), Equals, 0)
}

func (s *CliSuite) TestDownMultiple(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	s.FromText(c, p, "scale", "hello=2", SimpleTemplate)

	containers := s.GetContainersByProject(c, p)
	c.Assert(len(containers), Equals, 2)

	s.FromText(c, p, "down", SimpleTemplate)

	containers = s.GetContainersByProject(c, p)
	c.Assert(len(containers), Equals, 0)
}
