package integration

import (
	"fmt"
	"strings"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestPs(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")

	_, out := s.FromTextCaptureOutput(c, p, "ps", SimpleTemplate)

	c.Assert(strings.Contains(out,
		fmt.Sprintf(`%s  sh       Up Less than a second`, name)),
		Equals, true, Commentf("%s", out))
}

func (s *CliSuite) TestPsQuiet(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	container := s.GetContainerByName(c, name)

	_, out := s.FromTextCaptureOutput(c, p, "ps", "-q", SimpleTemplate)

	c.Assert(strings.Contains(out,
		fmt.Sprintf(`%s`, container.ID)),
		Equals, true, Commentf("%s", out))
}
