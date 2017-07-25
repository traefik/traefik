package credentials

import (
	"os/exec"

	"github.com/docker/cli/cli/config/configfile"
)

// DetectDefaultStore sets the default credentials store
// if the host includes the default store helper program.
func DetectDefaultStore(c *configfile.ConfigFile) {
	if c.CredentialsStore != "" {
		// user defined
		return
	}

	if defaultCredentialsStore != "" {
		if _, err := exec.LookPath(remoteCredentialsPrefix + defaultCredentialsStore); err == nil {
			c.CredentialsStore = defaultCredentialsStore
		}
	}
}
