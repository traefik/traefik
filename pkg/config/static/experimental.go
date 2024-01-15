package static

import "github.com/traefik/traefik/v3/pkg/plugins"

// Experimental experimental Traefik features.
type Experimental struct {
	Plugins      map[string]plugins.Descriptor      `description:"Plugins configuration." json:"plugins,omitempty" toml:"plugins,omitempty" yaml:"plugins,omitempty" export:"true"`
	LocalPlugins map[string]plugins.LocalDescriptor `description:"Local plugins configuration." json:"localPlugins,omitempty" toml:"localPlugins,omitempty" yaml:"localPlugins,omitempty" export:"true"`

	KubernetesGateway bool `description:"Allow the Kubernetes gateway api provider usage." json:"kubernetesGateway,omitempty" toml:"kubernetesGateway,omitempty" yaml:"kubernetesGateway,omitempty" export:"true"`
}
