package dynamic

// +k8s:deepcopy-gen=true

// TCPMiddleware holds the TCPMiddleware configuration.
type TCPMiddleware struct {
	InFlightConn *TCPInFlightConn `json:"inFlightConn,omitempty" toml:"inFlightConn,omitempty" yaml:"inFlightConn,omitempty" export:"true"`
	// Deprecated: please use IPAllowList instead.
	IPWhiteList *TCPIPWhiteList `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty" export:"true"`
	IPAllowList *TCPIPAllowList `json:"ipAllowList,omitempty" toml:"ipAllowList,omitempty" yaml:"ipAllowList,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPInFlightConn holds the TCP InFlightConn middleware configuration.
// This middleware prevents services from being overwhelmed with high load,
// by limiting the number of allowed simultaneous connections for one IP.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/tcp/inflightconn/
type TCPInFlightConn struct {
	// Amount defines the maximum amount of allowed simultaneous connections.
	// The middleware closes the connection if there are already amount connections opened.
	// +kubebuilder:validation:Minimum=0
	Amount int64 `json:"amount,omitempty" toml:"amount,omitempty" yaml:"amount,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPIPWhiteList holds the TCP IPWhiteList middleware configuration.
// Deprecated: please use IPAllowList instead.
type TCPIPWhiteList struct {
	// SourceRange defines the allowed IPs (or ranges of allowed IPs by using CIDR notation).
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}

// +k8s:deepcopy-gen=true

// TCPIPAllowList holds the TCP IPAllowList middleware configuration.
// This middleware limits allowed requests based on the client IP.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/tcp/ipallowlist/
type TCPIPAllowList struct {
	// SourceRange defines the allowed IPs (or ranges of allowed IPs by using CIDR notation).
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}
