package integration

import (
	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestPull(c *C) {
	//TODO: This doesn't test much
	s.ProjectFromText(c, "pull", `
        hello:
          image: tianon/true
          stdin_open: true
          tty: true
        `)
}
