package auth

import (
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/registry"
)

// Lookup defines a method for looking up authentication information
type Lookup interface {
	All() map[string]types.AuthConfig
	Lookup(repoInfo *registry.RepositoryInfo) types.AuthConfig
}

// ConfigLookup implements AuthLookup by reading a Docker config file
type ConfigLookup struct {
	*configfile.ConfigFile
}

// NewConfigLookup creates a new ConfigLookup for a given context
func NewConfigLookup(configfile *configfile.ConfigFile) *ConfigLookup {
	return &ConfigLookup{
		ConfigFile: configfile,
	}
}

// Lookup uses a Docker config file to lookup authentication information
func (c *ConfigLookup) Lookup(repoInfo *registry.RepositoryInfo) types.AuthConfig {
	if c.ConfigFile == nil || repoInfo == nil || repoInfo.Index == nil {
		return types.AuthConfig{}
	}
	return registry.ResolveAuthConfig(c.ConfigFile.AuthConfigs, repoInfo.Index)
}

// All uses a Docker config file to get all authentication information
func (c *ConfigLookup) All() map[string]types.AuthConfig {
	if c.ConfigFile == nil {
		return map[string]types.AuthConfig{}
	}
	return c.ConfigFile.AuthConfigs
}
