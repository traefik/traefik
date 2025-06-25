package static

import "github.com/traefik/traefik/v3/pkg/plugins"

// Experimental experimental Traefik features.
type Experimental struct {
	Plugins                map[string]plugins.Descriptor      `description:"Plugins configuration." json:"plugins,omitempty" toml:"plugins,omitempty" yaml:"plugins,omitempty" export:"true"`
	LocalPlugins           map[string]plugins.LocalDescriptor `description:"Local plugins configuration." json:"localPlugins,omitempty" toml:"localPlugins,omitempty" yaml:"localPlugins,omitempty" export:"true"`
	AbortOnPluginFailure   bool                               `description:"Defines whether all plugins must be loaded successfully for Traefik to start." json:"abortOnPluginFailure,omitempty" toml:"abortOnPluginFailure,omitempty" yaml:"abortOnPluginFailure,omitempty" export:"true"`
	FastProxy              *FastProxyConfig                   `description:"Enables the FastProxy implementation." json:"fastProxy,omitempty" toml:"fastProxy,omitempty" yaml:"fastProxy,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	OTLPLogs               bool                               `description:"Enables the OpenTelemetry logs integration." json:"otlplogs,omitempty" toml:"otlplogs,omitempty" yaml:"otlplogs,omitempty" export:"true"`
	KubernetesIngressNGINX bool                               `description:"Allow the Kubernetes Ingress NGINX provider usage." json:"kubernetesIngressNGINX,omitempty" toml:"kubernetesIngressNGINX,omitempty" yaml:"kubernetesIngressNGINX,omitempty" export:"true"`

	// Deprecated: KubernetesGateway provider is not an experimental feature starting with v3.1. Please remove its usage from the static configuration.
	KubernetesGateway bool `description:"(Deprecated) Allow the Kubernetes gateway api provider usage." json:"kubernetesGateway,omitempty" toml:"kubernetesGateway,omitempty" yaml:"kubernetesGateway,omitempty" export:"true"`
}

// FastProxyConfig holds the FastProxy configuration.
type FastProxyConfig struct {
	Debug bool `description:"Enable debug mode for the FastProxy implementation." json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty" export:"true"`
}
