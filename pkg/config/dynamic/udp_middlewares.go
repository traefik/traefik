package dynamic

// +k8s:deepcopy-gen=true

// UDPMiddleware holds the UDPMiddleware configuration.
type UDPMiddleware struct {
	IPWhiteList *UDPIPWhiteList `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// UDPIPWhiteList holds the UDP ip white list configuration.
type UDPIPWhiteList struct {
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}
