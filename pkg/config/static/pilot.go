package static

// Pilot Configuration related to Traefik Pilot.
// Deprecated.
type Pilot struct {
	Token     string `description:"Traefik Pilot token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	Dashboard bool   `description:"Enable Traefik Pilot in the dashboard." json:"dashboard,omitempty" toml:"dashboard,omitempty" yaml:"dashboard,omitempty"`
}

// SetDefaults sets the default values.
func (p *Pilot) SetDefaults() {
	p.Dashboard = true
}
