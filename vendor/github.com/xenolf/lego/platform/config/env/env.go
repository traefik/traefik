package env

import (
	"fmt"
	"os"
	"strconv"
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

// GetOrDefaultInt returns the given environment variable value as an integer.
// Returns the default if the envvar cannot be coopered to an int, or is not found.
func GetOrDefaultInt(envVar string, defaultValue int) int {
	v, err := strconv.Atoi(os.Getenv(envVar))
	if err != nil {
		return defaultValue
	}

	return v
}
