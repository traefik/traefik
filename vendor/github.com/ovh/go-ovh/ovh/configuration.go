package ovh

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// Use variables for easier test overload
var (
	systemConfigPath = "/etc/ovh.conf"
	userConfigPath   = "/.ovh.conf" // prefixed with homeDir
	localConfigPath  = "./ovh.conf"
)

// currentUserHome attempts to get current user's home directory
func currentUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

// appendConfigurationFile only if it exists. We need to do this because
// ini package will fail to load configuration at all if a configuration
// file is missing. This is racy, but better than always failing.
func appendConfigurationFile(cfg *ini.File, path string) {
	if file, err := os.Open(path); err == nil {
		file.Close()
		cfg.Append(path)
	}
}

// loadConfig loads client configuration from params, environments or configuration
// files (by order of decreasing precedence).
//
// loadConfig will check OVH_CONSUMER_KEY, OVH_APPLICATION_KEY, OVH_APPLICATION_SECRET
// and OVH_ENDPOINT environment variables. If any is present, it will take precedence
// over any configuration from file.
//
// Configuration files are ini files. They share the same format as python-ovh,
// node-ovh, php-ovh and all other wrappers. If any wrapper is configured, all
// can re-use the same configuration. loadConfig will check for configuration in:
//
// - ./ovh.conf
// - $HOME/.ovh.conf
// - /etc/ovh.conf
//
func (c *Client) loadConfig(endpointName string) error {
	// Load configuration files by order of increasing priority. All configuration
	// files are optional. Only load file from user home if home could be resolve
	cfg := ini.Empty()
	appendConfigurationFile(cfg, systemConfigPath)
	if home, err := currentUserHome(); err == nil {
		userConfigFullPath := filepath.Join(home, userConfigPath)
		appendConfigurationFile(cfg, userConfigFullPath)
	}
	appendConfigurationFile(cfg, localConfigPath)

	// Canonicalize configuration
	if endpointName == "" {
		endpointName = getConfigValue(cfg, "default", "endpoint", "ovh-eu")
	}

	if c.AppKey == "" {
		c.AppKey = getConfigValue(cfg, endpointName, "application_key", "")
	}

	if c.AppSecret == "" {
		c.AppSecret = getConfigValue(cfg, endpointName, "application_secret", "")
	}

	if c.ConsumerKey == "" {
		c.ConsumerKey = getConfigValue(cfg, endpointName, "consumer_key", "")
	}

	// Load real endpoint URL by name. If endpoint contains a '/', consider it as a URL
	if strings.Contains(endpointName, "/") {
		c.endpoint = endpointName
	} else {
		c.endpoint = Endpoints[endpointName]
	}

	// If we still have no valid endpoint, AppKey or AppSecret, return an error
	if c.endpoint == "" {
		return fmt.Errorf("Unknown endpoint '%s'. Consider checking 'Endpoints' list of using an URL.", endpointName)
	}
	if c.AppKey == "" {
		return fmt.Errorf("Missing application key. Please check your configuration or consult the documentation to create one.")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("Missing application secret. Please check your configuration or consult the documentation to create one.")
	}

	return nil
}

// getConfigValue returns the value of OVH_<NAME> or ``name`` value from ``section``. If
// the value could not be read from either env or any configuration files, return 'def'
func getConfigValue(cfg *ini.File, section, name, def string) string {
	// Attempt to load from environment
	fromEnv := os.Getenv("OVH_" + strings.ToUpper(name))
	if len(fromEnv) > 0 {
		return fromEnv
	}

	// Attempt to load from configuration
	fromSection := cfg.Section(section)
	if fromSection == nil {
		return def
	}

	fromSectionKey := fromSection.Key(name)
	if fromSectionKey == nil {
		return def
	}
	return fromSectionKey.String()
}
