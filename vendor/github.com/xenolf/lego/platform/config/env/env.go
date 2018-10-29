package env

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xenolf/lego/log"
)

// Get environment variables
func Get(names ...string) (map[string]string, error) {
	values := map[string]string{}

	var missingEnvVars []string
	for _, envVar := range names {
		value := GetOrFile(envVar)
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

// GetWithFallback Get environment variable values
// The first name in each group is use as key in the result map
//
//	// LEGO_ONE="ONE"
//	// LEGO_TWO="TWO"
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => "LEGO_ONE" = "ONE"
//
// ----
//
//	// LEGO_ONE=""
//	// LEGO_TWO="TWO"
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => "LEGO_ONE" = "TWO"
//
// ----
//
//	// LEGO_ONE=""
//	// LEGO_TWO=""
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => error
//
func GetWithFallback(groups ...[]string) (map[string]string, error) {
	values := map[string]string{}

	var missingEnvVars []string
	for _, names := range groups {
		if len(names) == 0 {
			return nil, errors.New("undefined environment variable names")
		}

		value, envVar := getOneWithFallback(names[0], names[1:]...)
		if len(value) == 0 {
			missingEnvVars = append(missingEnvVars, envVar)
			continue
		}
		values[envVar] = value
	}

	if len(missingEnvVars) > 0 {
		return nil, fmt.Errorf("some credentials information are missing: %s", strings.Join(missingEnvVars, ","))
	}

	return values, nil
}

func getOneWithFallback(main string, names ...string) (string, string) {
	value := GetOrFile(main)
	if len(value) > 0 {
		return value, main
	}

	for _, name := range names {
		value := GetOrFile(name)
		if len(value) > 0 {
			return value, main
		}
	}

	return "", main
}

// GetOrDefaultInt returns the given environment variable value as an integer.
// Returns the default if the envvar cannot be coopered to an int, or is not found.
func GetOrDefaultInt(envVar string, defaultValue int) int {
	v, err := strconv.Atoi(GetOrFile(envVar))
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
	v := GetOrFile(envVar)
	if len(v) == 0 {
		return defaultValue
	}

	return v
}

// GetOrDefaultBool returns the given environment variable value as a boolean.
// Returns the default if the envvar cannot be coopered to a boolean, or is not found.
func GetOrDefaultBool(envVar string, defaultValue bool) bool {
	v, err := strconv.ParseBool(GetOrFile(envVar))
	if err != nil {
		return defaultValue
	}

	return v
}

// GetOrFile Attempts to resolve 'key' as an environment variable.
// Failing that, it will check to see if '<key>_FILE' exists.
// If so, it will attempt to read from the referenced file to populate a value.
func GetOrFile(envVar string) string {
	envVarValue := os.Getenv(envVar)
	if envVarValue != "" {
		return envVarValue
	}

	fileVar := envVar + "_FILE"
	fileVarValue := os.Getenv(fileVar)
	if fileVarValue == "" {
		return envVarValue
	}

	fileContents, err := ioutil.ReadFile(fileVarValue)
	if err != nil {
		log.Printf("Failed to read the file %s (defined by env var %s): %s", fileVarValue, fileVar, err)
		return ""
	}

	return string(fileContents)
}
