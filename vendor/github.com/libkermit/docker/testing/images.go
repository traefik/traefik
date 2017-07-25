package testing

import (
	"testing"
)

// Pull pulls the given reference (image)
func (p *Project) Pull(t *testing.T, ref string) {
	if err := p.project.Pull(ref); err != nil {
		t.Fatalf("Error while pulling image %s: %s", ref, err.Error())
	}
}
