package runtimeconfig

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"gopkg.in/yaml.v3"
)

func TestResolve(t *testing.T) {
	type auth struct{ Token string }
	type credential struct{ Auth auth }
	type target struct {
		URL     string
		Headers map[string]string
	}
	type rule struct{ HTTP *target }
	config := struct {
		RuntimeValues  map[string]dynamic.RuntimeValue
		CredentialSets map[string]credential
		Rules          []rule
	}{
		RuntimeValues: map[string]dynamic.RuntimeValue{
			"token.any-name": {Source: "secret:apps/action:key", Value: "secret-value"},
			"host":           {Source: "configmap:apps/action:host", Value: "example.test"},
		},
		CredentialSets: map[string]credential{
			"arbitrary": {Auth: auth{Token: "runtime://token.any-name"}},
		},
		Rules: []rule{{
			HTTP: &target{
				URL:     "https://${runtime.host}/events",
				Headers: map[string]string{"X-Token": "${runtime.token.any-name}"},
			},
		}},
	}

	require.NoError(t, Resolve(config.RuntimeValues, &config))
	require.Equal(t, "secret-value", config.CredentialSets["arbitrary"].Auth.Token)
	require.Equal(t, "https://example.test/events", config.Rules[0].HTTP.URL)
	require.Equal(t, "secret-value", config.Rules[0].HTTP.Headers["X-Token"])
	require.Equal(t, "secret-value", config.RuntimeValues["token.any-name"].Value)
}

func TestResolveRejectsUnknownValue(t *testing.T) {
	config := struct {
		Value string
	}{Value: "runtime://missing"}
	require.EqualError(t, Resolve(nil, &config), `runtime value "missing" is not declared`)
}

func TestSourceIdentityDoesNotContainMaterial(t *testing.T) {
	values := map[string]dynamic.RuntimeValue{
		"dsn": {Source: "secret:apps/audit:connection", Value: "user:password@tcp(db)/audit"},
	}
	identity, err := SourceIdentity("runtime://dsn", "", values)
	require.NoError(t, err)
	require.Equal(t, "runtime:secret:apps/audit:connection", identity)
	require.NotContains(t, identity, "password")
}

func TestRuntimeMaterialIsNotSerialized(t *testing.T) {
	value := dynamic.RuntimeValue{
		Source:   "secret:apps/audit:connection",
		Revision: "1234",
		Value:    "do-not-publish",
	}

	jsonValue, err := json.Marshal(value)
	require.NoError(t, err)
	require.NotContains(t, string(jsonValue), value.Value)

	yamlValue, err := yaml.Marshal(value)
	require.NoError(t, err)
	require.NotContains(t, string(yamlValue), value.Value)

	var tomlValue strings.Builder
	require.NoError(t, toml.NewEncoder(&tomlValue).Encode(value))
	require.NotContains(t, tomlValue.String(), value.Value)
}
