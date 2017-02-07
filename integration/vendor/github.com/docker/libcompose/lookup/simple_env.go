package lookup

import (
	"fmt"
	"os"

	"github.com/docker/libcompose/config"
)

// OsEnvLookup is a "bare" structure that implements the project.EnvironmentLookup interface
type OsEnvLookup struct {
}

// Lookup creates a string slice of string containing a "docker-friendly" environment string
// in the form of 'key=value'. It gets environment values using os.Getenv.
// If the os environment variable does not exists, the slice is empty. serviceName and config
// are not used at all in this implementation.
func (o *OsEnvLookup) Lookup(key, serviceName string, config *config.ServiceConfig) []string {
	ret := os.Getenv(key)
	if ret == "" {
		return []string{}
	}
	return []string{fmt.Sprintf("%s=%s", key, ret)}
}
