package lookup

import (
	"testing"
)

func TestOsEnvLookup(t *testing.T) {
	osEnvLookup := &OsEnvLookup{}

	envs := osEnvLookup.Lookup("PATH", nil)
	if len(envs) != 1 {
		t.Fatalf("Expected envs to contains one element, but was %v", envs)
	}

	envs = osEnvLookup.Lookup("path", nil)
	if len(envs) != 0 {
		t.Fatalf("Expected envs to be empty, but was %v", envs)
	}

	envs = osEnvLookup.Lookup("DOES_NOT_EXIST", nil)
	if len(envs) != 0 {
		t.Fatalf("Expected envs to be empty, but was %v", envs)
	}
}
