package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestRestart(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(cn.State.Running, Equals, true)
	time := cn.State.StartedAt

	s.FromText(c, p, "restart", SimpleTemplate)

	cn = s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)
	c.Assert(cn.State.Running, Equals, true)

	c.Assert(time, Not(Equals), cn.State.StartedAt)
}
