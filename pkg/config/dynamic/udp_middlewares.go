package dynamic

// +k8s:deepcopy-gen=true

// UDPMiddleware holds the UDPMiddleware configuration.
type UDPMiddleware struct {
	IPAllowList *UDPIPAllowList `json:"ipAllowList,omitempty" toml:"ipAllowList,omitempty" yaml:"ipAllowList,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// UDPIPAllowList holds the UDP ip allow list configuration.
type UDPIPAllowList struct {
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}
