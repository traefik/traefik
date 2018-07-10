package env

import (
	"fmt"
	"os"
	"strings"
)

// Get environment variables
func Get(names ...string) (map[string]string, error) {
	values := map[string]string{}

	var missingEnvVars []string
	for _, envVar := range names {
		value := os.Getenv(envVar)
		if value == "" {
			missingEnvVars = append(missingEnvVars, envVar)
		}
		values[envVar] = value
	}

	if len(missingEnvVars) > 0 {
		return nil, fmt.Errorf("some credentials information are missing: %s", strings.Join(missingEnvVars, ","))
	}

	return values, nil
}
