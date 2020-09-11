package static

// Pilot Configuration related to Traefik Pilot.
type Pilot struct {
	Token string `description:"Traefik Pilot token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty"`
}
