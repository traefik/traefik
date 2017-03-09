package docker

import (
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/types"
)

// AuthLookup defines a method for looking up authentication information
type AuthLookup interface {
	All() map[string]types.AuthConfig
	Lookup(repoInfo *registry.RepositoryInfo) types.AuthConfig
}

// ConfigAuthLookup implements AuthLookup by reading a Docker config file
type ConfigAuthLookup struct {
	context *Context
}

// NewConfigAuthLookup creates a new ConfigAuthLookup for a given context
func NewConfigAuthLookup(context *Context) *ConfigAuthLookup {
	return &ConfigAuthLookup{
		context: context,
	}
}

// Lookup uses a Docker config file to lookup authentication information
func (c *ConfigAuthLookup) Lookup(repoInfo *registry.RepositoryInfo) types.AuthConfig {
	if c.context.ConfigFile == nil || repoInfo == nil || repoInfo.Index == nil {
		return types.AuthConfig{}
	}
	return registry.ResolveAuthConfig(c.context.ConfigFile.AuthConfigs, repoInfo.Index)
}

// All uses a Docker config file to get all authentication information
func (c *ConfigAuthLookup) All() map[string]types.AuthConfig {
	if c.context.ConfigFile == nil {
		return map[string]types.AuthConfig{}
	}
	return c.context.ConfigFile.AuthConfigs
}
