package credentials

import (
	"os/exec"
)

// DetectDefaultStore return the default credentials store for the platform if
// the store executable is available.
func DetectDefaultStore(store string) string {
	platformDefault := defaultCredentialsStore()

	// user defined or no default for platform
	if store != "" || platformDefault == "" {
		return store
	}

	if _, err := exec.LookPath(remoteCredentialsPrefix + platformDefault); err == nil {
		return platformDefault
	}
	return ""
}
