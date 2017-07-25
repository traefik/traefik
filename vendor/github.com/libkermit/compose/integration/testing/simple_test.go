package composeit

import (
	"testing"

	compose "github.com/libkermit/compose/testing"
)

func TestSimpleProject(t *testing.T) {
	project := compose.CreateProject(t, "simple", "../assets/simple.yml")
	project.Start(t)

	container := project.Container(t, "hello")
	if container.Name != "/simple_hello_1" {
		t.Fatalf("expected name '/simple_hello_1', got %s", container.Name)
	}

	project.Stop(t)
}
