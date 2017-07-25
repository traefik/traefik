package dockerit

import (
	"strings"
	"testing"
)

func TestCreateSimple(t *testing.T) {
	project := setupTest(t)

	container := project.Create(t, "busybox")

	if container.ID == "" {
		t.Fatalf("expected a containerId, got nothing")
	}
	if strings.HasPrefix(container.Name, "kermit_") {
		t.Fatalf("expected name to start with 'kermit_', got %s", container.Name)
	}
}

func TestStartAndStop(t *testing.T) {
	project := setupTest(t)

	container := project.Start(t, "busybox")

	if container.ID == "" {
		t.Fatalf("expected a containerId, got nothing")
	}
	if strings.HasPrefix(container.Name, "kermit_") {
		t.Fatalf("expected name to start with 'kermit_', got %s", container.Name)
	}
	if !container.State.Running {
		t.Fatalf("expected container to be running, but was in state %v", container.State)
	}

	project.Stop(t, container.ID)

	container = project.Inspect(t, container.ID)
	if container.State.Running {
		t.Fatalf("expected container to be running, but was in state %v", container.State)
	}
}
