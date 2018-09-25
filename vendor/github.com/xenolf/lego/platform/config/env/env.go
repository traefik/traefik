package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

// GetOrDefaultSecond returns the given environment variable value as an time.Duration (second).
// Returns the default if the envvar cannot be coopered to an int, or is not found.
func GetOrDefaultSecond(envVar string, defaultValue time.Duration) time.Duration {
	v := GetOrDefaultInt(envVar, -1)
	if v < 0 {
		return defaultValue
	}

	return time.Duration(v) * time.Second
}

// GetOrDefaultString returns the given environment variable value as a string.
// Returns the default if the envvar cannot be find.
func GetOrDefaultString(envVar string, defaultValue string) string {
	v := os.Getenv(envVar)
	if len(v) == 0 {
		return defaultValue
	}

	return v
}

// GetOrDefaultBool returns the given environment variable value as a boolean.
// Returns the default if the envvar cannot be coopered to a boolean, or is not found.
func GetOrDefaultBool(envVar string, defaultValue bool) bool {
	v, err := strconv.ParseBool(os.Getenv(envVar))
	if err != nil {
		return defaultValue
	}

	return v
}
