package static

// Pilot Configuration related to Traefik Pilot.
type Pilot struct {
	Token string `description:"Traefik Pilot token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty"`
	Private bool `description:"Traefik Pilot private enabled." json:"private,omitempty" toml:"private,omitempty" yaml:"private,omitempty"`
	RepoUrl string `description:"Traefik Pilot private repository url." json:"repo,omitempty" toml:"repo,omitempty" yaml:"repo,omitempty"`
}

// SetDefaults sets the default values.
func (a *Pilot) SetDefaults() {
    a.RepoUrl = "https://plugin.pilot.traefik.io/public/"
}

