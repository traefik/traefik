package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestCreateWithEnvInCurrentDir(c *C) {
	cwd, err := os.Getwd()
	c.Assert(err, IsNil)
	defer os.Chdir(cwd)

	c.Assert(os.Chdir("./assets/env"), IsNil, Commentf("Could not change current directory to ./assets/env"))

	projectName := s.RandomProject()
	cmd := exec.Command("../../../bundles/libcompose-cli", "--verbose", "-p", projectName, "-f", "-", "create")
	cmd.Stdin = bytes.NewBufferString(`
hello:
  image: tianon/true
  labels:
    - "FOO=${FOO}"
`)
	output, err := cmd.CombinedOutput()
	c.Assert(err, IsNil, Commentf("%s", output))

	name := fmt.Sprintf("%s_%s_1", projectName, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(len(cn.Config.Labels), Equals, 7, Commentf("%v", cn.Config.Env))
	c.Assert(cn.Config.Labels["FOO"], Equals, "bar", Commentf("%v", cn.Config.Labels))
}

func (s *CliSuite) TestCreateWithEnvNotInCurrentDir(c *C) {
	p := s.CreateProjectFromText(c, `
hello:
  image: tianon/true
`)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(len(cn.Config.Labels), Equals, 6, Commentf("%v", cn.Config.Labels))
}
