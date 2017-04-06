package lookup

import (
	"github.com/docker/libcompose/config"
)

// ComposableEnvLookup is a structure that implements the project.EnvironmentLookup interface.
// It holds an ordered list of EnvironmentLookup to call to look for the environment value.
type ComposableEnvLookup struct {
	Lookups []config.EnvironmentLookup
}

// Lookup creates a string slice of string containing a "docker-friendly" environment string
// in the form of 'key=value'. It loop through the lookups and returns the latest value if
// more than one lookup return a result.
func (l *ComposableEnvLookup) Lookup(key, serviceName string, config *config.ServiceConfig) []string {
	result := []string{}
	for _, lookup := range l.Lookups {
		env := lookup.Lookup(key, serviceName, config)
		if len(env) == 1 {
			result = env
		}
	}
	return result
}
