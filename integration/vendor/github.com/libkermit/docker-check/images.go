package docker

import (
	"github.com/go-check/check"
)

// Pull pulls the given reference (image)
func (p *Project) Pull(c *check.C, ref string) {
	c.Assert(p.project.Pull(ref), check.IsNil,
		check.Commentf("Error while pulling image %s: %s", ref))
}
