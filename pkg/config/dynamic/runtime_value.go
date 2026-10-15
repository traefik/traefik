package dynamic

// +k8s:deepcopy-gen=true

// RuntimeValue is a named, provider-resolved runtime value.
//
// Source and Revision are safe metadata used to identify and reload the value.
// Value is deliberately excluded from JSON so secrets cannot be exposed by the
// dashboard, API, or dynamic-configuration debug logs. File is a local source
// supported by file-based configurations and is watched by runtime consumers.
type RuntimeValue struct {
	Source   string `json:"source,omitempty" toml:"source,omitempty" yaml:"source,omitempty"`
	Revision string `json:"revision,omitempty" toml:"revision,omitempty" yaml:"revision,omitempty"`
	File     string `json:"file,omitempty" toml:"file,omitempty" yaml:"file,omitempty"`
	Value    string `json:"-" toml:"-" yaml:"-"`
}
