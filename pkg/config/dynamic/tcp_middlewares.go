package dynamic

import (
	"time"

	ptypes "github.com/traefik/paerser/types"
)

// +k8s:deepcopy-gen=true

// TCPMiddleware holds the TCPMiddleware configuration.
type TCPMiddleware struct {
	IPWhiteList *TCPIPWhiteList `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty" export:"true"`
	RateLimit   *TCPRateLimit   `json:"rateLimit,omitempty" toml:"rateLimit,omitempty" yaml:"rateLimit,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPIPWhiteList holds the TCP ip white list configuration.
type TCPIPWhiteList struct {
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}

// +k8s:deepcopy-gen=true

// TCPRateLimit holds the rate limiting configuration for a given router.
type TCPRateLimit struct {
	// Average is the maximum rate, by default in requests/s, allowed for the given source.
	// It defaults to 0, which means no rate limiting.
	// The rate is actually defined by dividing Average by Period. So for a rate below 1req/s,
	// one needs to define a Period larger than a second.
	Average int64 `json:"average,omitempty" toml:"average,omitempty" yaml:"average,omitempty" export:"true"`

	// Period, in combination with Average, defines the actual maximum rate, such as:
	// r = Average / Period. It defaults to a second.
	Period ptypes.Duration `json:"period,omitempty" toml:"period,omitempty" yaml:"period,omitempty" export:"true"`

	// Burst is the maximum number of requests allowed to arrive in the same arbitrarily small period of time.
	// It defaults to 1.
	Burst int64 `json:"burst,omitempty" toml:"burst,omitempty" yaml:"burst,omitempty" export:"true"`
}

// SetDefaults sets the default values on a RateLimit.
func (r *TCPRateLimit) SetDefaults() {
	r.Burst = 1
	r.Period = ptypes.Duration(time.Second)
}
