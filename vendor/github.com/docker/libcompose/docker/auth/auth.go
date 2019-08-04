package auth

import (
	"github.com/docker/cli/cli/config/configfile"
	clitypes "github.com/docker/cli/cli/config/types"
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
	return registry.ResolveAuthConfig(convert(c.ConfigFile.AuthConfigs), repoInfo.Index)
}

// All uses a Docker config file to get all authentication information
func (c *ConfigLookup) All() map[string]types.AuthConfig {
	if c.ConfigFile == nil {
		return map[string]types.AuthConfig{}
	}
	return convert(c.ConfigFile.AuthConfigs)
}

func convert(acs map[string]clitypes.AuthConfig) map[string]types.AuthConfig {
	if acs == nil {
		return nil
	}

	result := map[string]types.AuthConfig{}
	for k, v := range acs {
		result[k] = types.AuthConfig{
			Username:      v.Username,
			Password:      v.Password,
			Auth:          v.Auth,
			Email:         v.Email,
			ServerAddress: v.ServerAddress,
			IdentityToken: v.IdentityToken,
			RegistryToken: v.RegistryToken,
		}
	}
	return result
}
