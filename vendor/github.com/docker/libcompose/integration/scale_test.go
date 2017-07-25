package integration

import (
	"fmt"
	"strings"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestScale(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	name2 := fmt.Sprintf("%s_%s_2", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(cn.State.Running, Equals, true)

	containers := s.GetContainersByProject(c, p)
	c.Assert(1, Equals, len(containers))

	s.FromText(c, p, "scale", "hello=2", SimpleTemplate)

	containers = s.GetContainersByProject(c, p)
	c.Assert(2, Equals, len(containers))

	for _, name := range []string{name, name2} {
		cn := s.GetContainerByName(c, name)
		c.Assert(cn, NotNil)
		c.Assert(cn.State.Running, Equals, true)
	}

	s.FromText(c, p, "scale", "--timeout", "0", "hello=1", SimpleTemplate)
	containers = s.GetContainersByProject(c, p)
	c.Assert(1, Equals, len(containers))

	cn = s.GetContainerByName(c, name)
	c.Assert(cn, IsNil)

	cn = s.GetContainerByName(c, name2)
	c.Assert(cn, NotNil)
	c.Assert(cn.State.Running, Equals, true)
}

func (s *CliSuite) TestScaleWithHostPortWarning(c *C) {
	template := `
	test:
	  image: busybox
	  ports:
	  - 8001:8001
	`
	p := s.ProjectFromText(c, "up", template)

	name := fmt.Sprintf("%s_%s_1", p, "test")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(cn.State.Running, Equals, true)

	containers := s.GetContainersByProject(c, p)
	c.Assert(1, Equals, len(containers))

	_, output := s.FromTextCaptureOutput(c, p, "scale", "test=2", template)

	// Assert warning is given when trying to scale a service that specifies a host port
	c.Assert(strings.Contains(output, "If multiple containers for this service are created on a single host, the port will clash."), Equals, true, Commentf(output))
}
