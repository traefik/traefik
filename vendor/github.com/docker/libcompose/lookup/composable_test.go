package lookup

import (
	"testing"

	"github.com/docker/libcompose/config"
)

type simpleEnvLookup struct {
	value []string
}

func (l *simpleEnvLookup) Lookup(key string, config *config.ServiceConfig) []string {
	return l.value
}

func TestComposableLookupWithoutAnyLookup(t *testing.T) {
	envLookup := &ComposableEnvLookup{}
	actuals := envLookup.Lookup("any", nil)
	if len(actuals) != 0 {
		t.Fatalf("expected an empty slice, got %v", actuals)
	}
}

func TestComposableLookupReturnsTheLastValue(t *testing.T) {
	envLookup1 := &simpleEnvLookup{
		value: []string{"value=1"},
	}
	envLookup2 := &simpleEnvLookup{
		value: []string{"value=2"},
	}
	envLookup := &ComposableEnvLookup{
		[]config.EnvironmentLookup{
			envLookup1,
			envLookup2,
		},
	}
	validateLookup(t, "value=2", envLookup.Lookup("value", nil))

	envLookup = &ComposableEnvLookup{
		[]config.EnvironmentLookup{
			envLookup2,
			envLookup1,
		},
	}
	validateLookup(t, "value=1", envLookup.Lookup("value", nil))
}
