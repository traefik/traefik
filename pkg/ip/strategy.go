package ip

import (
	"net"
	"net/http"
	"net/netip"
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
type RemoteAddrStrategy struct {
	// IPv6Subnet instructs the strategy to return the first IP of the subnet where IP belongs.
	IPv6Subnet *int
}

// GetIP returns the selected IP.
func (s *RemoteAddrStrategy) GetIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}

	if s.IPv6Subnet != nil {
		return getIPv6SubnetIP(ip, *s.IPv6Subnet)
	}

	return ip
}

// DepthStrategy a strategy based on the depth inside the X-Forwarded-For from right to left.
type DepthStrategy struct {
	Depth int
	// IPv6Subnet instructs the strategy to return the first IP of the subnet where IP belongs.
	IPv6Subnet *int
}

// GetIP returns the selected IP.
func (s *DepthStrategy) GetIP(req *http.Request) string {
	xff := req.Header.Get(xForwardedFor)
	xffs := strings.Split(xff, ",")

	if len(xffs) < s.Depth {
		return ""
	}

	ip := strings.TrimSpace(xffs[len(xffs)-s.Depth])

	if s.IPv6Subnet != nil {
		return getIPv6SubnetIP(ip, *s.IPv6Subnet)
	}

	return ip
}

// PoolStrategy is a strategy based on an IP Checker.
// It allows to check whether addresses are in a given pool of IPs.
type PoolStrategy struct {
	Checker *Checker
}

// GetIP checks the list of Forwarded IPs (most recent first) against the
// Checker pool of IPs. It returns the first IP that is not in the pool, or the
// empty string otherwise.
func (s *PoolStrategy) GetIP(req *http.Request) string {
	if s.Checker == nil {
		return ""
	}

	xff := req.Header.Get(xForwardedFor)
	xffs := strings.Split(xff, ",")

	for i := len(xffs) - 1; i >= 0; i-- {
		xffTrimmed := strings.TrimSpace(xffs[i])
		if len(xffTrimmed) == 0 {
			continue
		}
		if contain, _ := s.Checker.Contains(xffTrimmed); !contain {
			return xffTrimmed
		}
	}

	return ""
}

// getIPv6SubnetIP returns the IPv6 subnet IP.
// It returns the original IP when it is not an IPv6, or if parsing the IP has failed with an error.
func getIPv6SubnetIP(ip string, ipv6Subnet int) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ip
	}

	if !addr.Is6() {
		return ip
	}

	prefix, err := addr.Prefix(ipv6Subnet)
	if err != nil {
		return ip
	}

	return prefix.Addr().String()
}
