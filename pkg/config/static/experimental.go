package static

import "github.com/containous/traefik/v2/pkg/plugins"

// Experimental experimental Traefik features.
type Experimental struct {
	Pilot *Pilot `description:"Traefik Pilot configuration." json:"pilot,omitempty" toml:"pilot,omitempty" yaml:"pilot,omitempty"`

	Plugins   map[string]plugins.Descriptor `description:"Plugins configuration." json:"plugins,omitempty" toml:"plugins,omitempty" yaml:"plugins,omitempty"`
	DevPlugin *plugins.DevPlugin            `description:"Dev plugin configuration." json:"devPlugin,omitempty" toml:"devPlugin,omitempty" yaml:"devPlugin,omitempty"`
}

// Pilot Configuration related to Traefik Pilot.
type Pilot struct {
	Token string `description:"Traefik Pilot token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty"`
}
