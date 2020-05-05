package ip

import (
	"net"
	"net/http"
	"strings"
)

const (
	xForwardedFor = "X-Forwarded-For"
)

// Strategy a strategy for IP selection.
type Strategy interface {
	GetIP(req *http.Request) string
}

// RemoteAddrStrategy a strategy that always return the remote address.
type RemoteAddrStrategy struct{}

// GetIP returns the selected IP.
func (s *RemoteAddrStrategy) GetIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}

// DepthStrategy a strategy based on the depth inside the X-Forwarded-For from right to left.
type DepthStrategy struct {
	Depth int
}

// GetIP return the selected IP.
func (s *DepthStrategy) GetIP(req *http.Request) string {
	xff := req.Header.Get(xForwardedFor)
	xffs := strings.Split(xff, ",")

	if len(xffs) < s.Depth {
		return ""
	}
	return strings.TrimSpace(xffs[len(xffs)-s.Depth])
}

// CheckerStrategy a strategy based on an IP Checker
// allows to check that addresses are in a trusted IPs.
type CheckerStrategy struct {
	Checker *Checker
}

// GetIP return the selected IP.
func (s *CheckerStrategy) GetIP(req *http.Request) string {
	if s.Checker == nil {
		return ""
	}

	xff := req.Header.Get(xForwardedFor)
	xffs := strings.Split(xff, ",")

	for i := len(xffs) - 1; i >= 0; i-- {
		xffTrimmed := strings.TrimSpace(xffs[i])
		if contain, _ := s.Checker.Contains(xffTrimmed); !contain {
			return xffTrimmed
		}
	}
	return ""
}
