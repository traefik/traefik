package composeit

import (
	"testing"

	"github.com/go-check/check"
	compose "github.com/libkermit/compose/check"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type CheckSuite struct{}

var _ = check.Suite(&CheckSuite{})

func (s *CheckSuite) TestSimpleProject(c *check.C) {
	project := compose.CreateProject(c, "simple", "./assets/simple.yml")
	project.Start(c)

	container := project.Container(c, "hello")
	c.Assert(container.Name, check.Equals, "/simple_hello_1")

	project.Stop(c)
}
