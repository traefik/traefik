package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestStart(c *C) {
	p := s.ProjectFromText(c, "create", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")

	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)
	c.Assert(cn.State.Running, Equals, false)

	s.FromText(c, p, "start", SimpleTemplate)

	cn = s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)
	c.Assert(cn.State.Running, Equals, true)
}
