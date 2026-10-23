package runtimeconfig

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestResolveRejectsExcessiveDepth(t *testing.T) {
	var nested any = "runtime://token"
	for range 1024 {
		nested = []any{nested}
	}

	err := Resolve(map[string]dynamic.RuntimeValue{
		"token": {Value: "private-material"},
	}, &nested)
	require.EqualError(t, err, "runtime configuration exceeds maximum nesting depth")
	require.NotContains(t, err.Error(), "private-material")
}

func TestResolveRejectsCycles(t *testing.T) {
	type node struct{ Next *node }
	pointerCycle := &node{}
	pointerCycle.Next = pointerCycle
	mapCycle := map[string]any{}
	mapCycle["self"] = mapCycle
	sliceCycle := make([]any, 1)
	sliceCycle[0] = sliceCycle
	var interfaceCycle any
	interfaceCycle = &interfaceCycle

	for name, value := range map[string]any{
		"pointer":   pointerCycle,
		"map":       &mapCycle,
		"slice":     &sliceCycle,
		"interface": &interfaceCycle,
	} {
		t.Run(name, func(t *testing.T) {
			require.EqualError(t, Resolve(nil, value), "runtime configuration exceeds maximum nesting depth")
		})
	}
}

func TestResolveDepthBoundary(t *testing.T) {
	values := map[string]string{"token": "resolved"}
	value := "runtime://token"
	require.NoError(t, resolveValue(reflect.ValueOf(&value).Elem(), values, maxRuntimeDepth))
	require.Equal(t, "resolved", value)

	value = "runtime://token"
	require.EqualError(t, resolveValue(reflect.ValueOf(&value).Elem(), values, maxRuntimeDepth+1),
		"runtime configuration exceeds maximum nesting depth")
	require.Equal(t, "runtime://token", value)
}

func TestResolveAcceptsSharedReferences(t *testing.T) {
	type entry struct{ Value string }
	shared := &entry{Value: "runtime://token"}
	config := struct {
		Entries [2]*entry
		Values  map[string]any
	}{
		Entries: [2]*entry{shared, shared},
		Values: map[string]any{
			"nested":  []any{"runtime://token"},
			"runtime": dynamic.RuntimeValue{Value: "runtime://unresolved"},
		},
	}
	require.NoError(t, Resolve(map[string]dynamic.RuntimeValue{"token": {Value: "resolved"}}, &config))
	require.Equal(t, "resolved", shared.Value)
	require.Equal(t, "resolved", config.Values["nested"].([]any)[0])
	require.Equal(t, dynamic.RuntimeValue{Value: "runtime://unresolved"}, config.Values["runtime"])
}

func TestSourceIdentityNormalizesInputs(t *testing.T) {
	values := map[string]dynamic.RuntimeValue{
		"dsn": {Source: "secret:apps/connection:dsn", Value: "test-dsn"},
	}
	for name, input := range map[string]struct{ value, file string }{
		"literal":   {value: "test-dsn"},
		"reference": {value: "runtime://dsn"},
		"file":      {file: "/etc/traefik/connection"},
	} {
		t.Run(name, func(t *testing.T) {
			expected, err := SourceIdentity(input.value, input.file, values)
			require.NoError(t, err)
			actual, err := SourceIdentity(" \t"+input.value+"\n", " \t"+input.file+"\n", values)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
	_, err := SourceIdentity(" test-dsn ", " /etc/traefik/connection ", values)
	require.EqualError(t, err, "connection must not define both dsn and dsnFile")
}
