package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestDelete(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")

	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)
	c.Assert(cn.State.Running, Equals, true)

	s.FromText(c, p, "stop", SimpleTemplate)
	s.FromText(c, p, "rm", "--force", SimpleTemplate)

	cn = s.GetContainerByName(c, name)
	c.Assert(cn, IsNil)
}

func (s *CliSuite) TestDeleteOnlyRemovesStopped(c *C) {
	projectTemplate := `
hello:
  image: busybox
  stdin_open: true
  tty: true
bye:
  image: busybox
  stdin_open: true
  tty: true
`

	p := s.ProjectFromText(c, "up", projectTemplate)

	helloName := fmt.Sprintf("%s_%s_1", p, "hello")
	byeName := fmt.Sprintf("%s_%s_1", p, "bye")

	helloContainer := s.GetContainerByName(c, helloName)
	c.Assert(helloContainer, NotNil)
	c.Assert(helloContainer.State.Running, Equals, true)

	byeContainer := s.GetContainerByName(c, byeName)
	c.Assert(byeContainer, NotNil)
	c.Assert(byeContainer.State.Running, Equals, true)

	s.FromText(c, p, "stop", "bye", projectTemplate)

	byeContainer = s.GetContainerByName(c, byeName)
	c.Assert(byeContainer, NotNil)
	c.Assert(byeContainer.State.Running, Equals, false)

	s.FromText(c, p, "rm", "--force", projectTemplate)

	byeContainer = s.GetContainerByName(c, byeName)
	c.Assert(byeContainer, IsNil)

	helloContainer = s.GetContainerByName(c, helloName)
	c.Assert(helloContainer, NotNil)
}

func (s *CliSuite) TestDeleteWithVol(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")

	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)
	c.Assert(cn.State.Running, Equals, true)

	s.FromText(c, p, "stop", SimpleTemplate)
	s.FromText(c, p, "rm", "--force", "-v", SimpleTemplate)

	cn = s.GetContainerByName(c, name)
	c.Assert(cn, IsNil)
}
