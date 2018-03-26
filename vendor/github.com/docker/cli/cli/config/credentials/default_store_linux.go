package credentials

import (
	"github.com/docker/docker-credential-helpers/pass"
)

func defaultCredentialsStore() string {
	if pass.PassInitialized {
		return "pass"
	}

	return "secretservice"
}
