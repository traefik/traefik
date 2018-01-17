package lookup

import (
	"strings"

	"github.com/docker/cli/opts"
	"github.com/docker/libcompose/config"
)

// EnvfileLookup is a structure that implements the project.EnvironmentLookup interface.
// It holds the path of the file where to lookup environment values.
type EnvfileLookup struct {
	Path string
}

// Lookup creates a string slice of string containing a "docker-friendly" environment string
// in the form of 'key=value'. It gets environment values using a '.env' file in the specified
// path.
func (l *EnvfileLookup) Lookup(key string, config *config.ServiceConfig) []string {
	envs, err := opts.ParseEnvFile(l.Path)
	if err != nil {
		return []string{}
	}
	for _, env := range envs {
		e := strings.Split(env, "=")
		if e[0] == key {
			return []string{env}
		}
	}
	return []string{}
}
