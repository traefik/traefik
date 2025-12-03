package observability

import (
	"fmt"
	"os"
)

func EnsureUserEnvVar() error {
	if os.Getenv("USER") == "" {
		if err := os.Setenv("USER", "baqup"); err != nil {
			return fmt.Errorf("could not set USER environment variable: %w", err)
		}
	}
	return nil
}
