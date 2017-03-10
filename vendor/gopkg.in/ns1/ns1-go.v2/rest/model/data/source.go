package data

// Config is a flat mapping where values are simple (no slices/maps).
type Config map[string]interface{}

// Source wraps an NS1 /data/sources resource
type Source struct {
	ID string `json:"id,omitempty"`

	// Human readable name of the source.
	Name string `json:"name"`

	Type   string `json:"sourcetype"`
	Config Config `json:"config,omitempty"`
	Status string `json:"status,omitempty"`

	Feeds []*Feed `json:"feeds,omitempty"`
}

// NewSource takes a name and type t.
func NewSource(name string, t string) *Source {
	return &Source{
		Name:   name,
		Type:   t,
		Config: Config{},
		Feeds:  []*Feed{},
	}
}
