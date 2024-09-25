package static

import "github.com/traefik/traefik/v3/pkg/plugins"

// Experimental experimental Traefik features.
type Experimental struct {
	Plugins      map[string]plugins.Descriptor      `description:"Plugins configuration." json:"plugins,omitempty" toml:"plugins,omitempty" yaml:"plugins,omitempty" export:"true"`
	LocalPlugins map[string]plugins.LocalDescriptor `description:"Local plugins configuration." json:"localPlugins,omitempty" toml:"localPlugins,omitempty" yaml:"localPlugins,omitempty" export:"true"`

	FastProxy *FastProxyConfig `description:"Enable the FastProxy implementation." json:"fastProxy,omitempty" toml:"fastProxy,omitempty" yaml:"fastProxy,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	// Deprecated: KubernetesGateway provider is not an experimental feature starting with v3.1. Please remove its usage from the static configuration.
	KubernetesGateway bool `description:"(Deprecated) Allow the Kubernetes gateway api provider usage." json:"kubernetesGateway,omitempty" toml:"kubernetesGateway,omitempty" yaml:"kubernetesGateway,omitempty" export:"true"`
}

// FastProxyConfig holds the FastProxy configuration.
type FastProxyConfig struct {
	Debug bool `description:"Enable debug mode for the FastProxy implementation." json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty" export:"true"`
}
