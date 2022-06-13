package dynamic

// +k8s:deepcopy-gen=true

// TCPMiddleware holds the TCPMiddleware configuration.
type TCPMiddleware struct {
	InFlightConn *TCPInFlightConn `json:"inFlightConn,omitempty" toml:"inFlightConn,omitempty" yaml:"inFlightConn,omitempty" export:"true"`
	IPWhiteList  *TCPIPWhiteList  `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// To proactively prevent services from being overwhelmed with high load, the number of allowed simultaneous connections by IP can be limited. More info: https://doc.traefik.io/traefik/middlewares/tcp/inflightconn/
type TCPInFlightConn struct {
	// The amount option defines the maximum amount of allowed simultaneous connections. The middleware closes the connection if there are already amount connections opened.
	Amount int64 `json:"amount,omitempty" toml:"amount,omitempty" yaml:"amount,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// IPWhitelist accepts / refuses connections based on the client IP.
type TCPIPWhiteList struct {
	// The sourceRange option sets the allowed IPs (or ranges of allowed IPs by using CIDR notation).
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}
