package state

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/consul/structs"
)

func TestStateStore_Autopilot(t *testing.T) {
	s := testStateStore(t)

	expected := &structs.AutopilotConfig{
		CleanupDeadServers: true,
	}

	if err := s.AutopilotSetConfig(0, expected); err != nil {
		t.Fatal(err)
	}

	idx, config, err := s.AutopilotConfig()
	if err != nil {
		t.Fatal(err)
	}
	if idx != 0 {
		t.Fatalf("bad: %d", idx)
	}
	if !reflect.DeepEqual(expected, config) {
		t.Fatalf("bad: %#v, %#v", expected, config)
	}
}

func TestStateStore_AutopilotCAS(t *testing.T) {
	s := testStateStore(t)

	expected := &structs.AutopilotConfig{
		CleanupDeadServers: true,
	}

	if err := s.AutopilotSetConfig(0, expected); err != nil {
		t.Fatal(err)
	}
	if err := s.AutopilotSetConfig(1, expected); err != nil {
		t.Fatal(err)
	}

	// Do a CAS with an index lower than the entry
	ok, err := s.AutopilotCASConfig(2, 0, &structs.AutopilotConfig{
		CleanupDeadServers: false,
	})
	if ok || err != nil {
		t.Fatalf("expected (false, nil), got: (%v, %#v)", ok, err)
	}

	// Check that the index is untouched and the entry
	// has not been updated.
	idx, config, err := s.AutopilotConfig()
	if err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Fatalf("bad: %d", idx)
	}
	if !config.CleanupDeadServers {
		t.Fatalf("bad: %#v", config)
	}

	// Do another CAS, this time with the correct index
	ok, err = s.AutopilotCASConfig(2, 1, &structs.AutopilotConfig{
		CleanupDeadServers: false,
	})
	if !ok || err != nil {
		t.Fatalf("expected (true, nil), got: (%v, %#v)", ok, err)
	}

	// Make sure the config was updated
	idx, config, err = s.AutopilotConfig()
	if err != nil {
		t.Fatal(err)
	}
	if idx != 2 {
		t.Fatalf("bad: %d", idx)
	}
	if config.CleanupDeadServers {
		t.Fatalf("bad: %#v", config)
	}
}
