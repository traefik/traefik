package dynamic

// +k8s:deepcopy-gen=true

// TCPMiddleware holds the TCPMiddleware configuration.
type TCPMiddleware struct {
	InFlightConn *TCPInFlightConn `json:"InFlightConn,omitempty" toml:"InFlightConn,omitempty" yaml:"InFlightConn,omitempty" export:"true"`
	IPWhiteList  *TCPIPWhiteList  `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPInFlightConn holds the TCP in flight connection configuration.
type TCPInFlightConn struct {
	Amount int64 `json:"amount,omitempty" toml:"amount,omitempty" yaml:"amount,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPIPWhiteList holds the TCP ip white list configuration.
type TCPIPWhiteList struct {
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}
