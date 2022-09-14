package static

// Pilot Configuration related to Traefik Pilot.
// Deprecated.
type Pilot struct {
	Token     string `description:"Traefik Pilot token. (Deprecated)" json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	Dashboard bool   `description:"Enable Traefik Pilot in the dashboard. (Deprecated)" json:"dashboard,omitempty" toml:"dashboard,omitempty" yaml:"dashboard,omitempty"`
}
