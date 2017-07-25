package dockerit

import (
	"strings"
	"testing"

	"github.com/go-check/check"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type CheckSuite struct{}

var _ = check.Suite(&CheckSuite{})

func (s *CheckSuite) TestCreateSimple(c *check.C) {
	project := setupTest(c)

	container := project.Create(c, "busybox")

	c.Assert(container.ID, check.Not(check.Equals), "")
	c.Assert(strings.HasPrefix(container.Name, "kermit_"), check.Not(check.Equals), true)
}

func (s *CheckSuite) TestStartAndStop(c *check.C) {
	project := setupTest(c)

	container := project.Start(c, "busybox")

	c.Assert(container.ID, check.Not(check.Equals), "")
	c.Assert(strings.HasPrefix(container.Name, "kermit_"), check.Not(check.Equals), true)
	c.Assert(container.State.Running, check.Equals, true,
		check.Commentf("expected container to be running, but was in state %v", container.State))

	project.Stop(c, container.ID)

	container = project.Inspect(c, container.ID)
	c.Assert(container.State.Running, check.Equals, false,
		check.Commentf("expected container to not be running, but was in state %v", container.State))
}
