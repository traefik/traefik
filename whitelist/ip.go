package whitelist

import (
	"fmt"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

const (
	// XForwardedFor Header name
	XForwardedFor = "X-Forwarded-For"
)

// IP allows to check that addresses are in a white list
type IP struct {
	whiteListsIPs    []*net.IP
	whiteListsNet    []*net.IPNet
	insecure         bool
	useXForwardedFor bool
}

// NewIP builds a new IP given a list of CIDR-Strings to white list
func NewIP(whiteList []string, insecure bool, useXForwardedFor bool) (*IP, error) {
	if len(whiteList) == 0 && !insecure {
		return nil, errors.New("no white list provided")
	}

	ip := IP{
		insecure:         insecure,
		useXForwardedFor: useXForwardedFor,
	}

	if !insecure {
		for _, ipMask := range whiteList {
			if ipAddr := net.ParseIP(ipMask); ipAddr != nil {
				ip.whiteListsIPs = append(ip.whiteListsIPs, &ipAddr)
			} else {
				_, ipAddr, err := net.ParseCIDR(ipMask)
				if err != nil {
					return nil, fmt.Errorf("parsing CIDR white list %s: %v", ipAddr, err)
				}
				ip.whiteListsNet = append(ip.whiteListsNet, ipAddr)
			}
		}
	}

	return &ip, nil
}

// IsAuthorized checks if provided request is authorized by the white list
func (ip *IP) IsAuthorized(req *http.Request) (bool, net.IP, error) {
	if ip.insecure {
		return true, nil, nil
	}

	if ip.useXForwardedFor {
		xFFs := req.Header[XForwardedFor]
		if ip.useXForwardedFor && len(xFFs) > 1 {
			for _, xFF := range xFFs {
				ok, i, err := ip.contains(parseHost(xFF))
				if err != nil {
					return false, nil, err
				}

				if ok {
					return ok, i, nil
				}
			}
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return false, nil, err
	}
	return ip.contains(host)
}

// contains checks if provided address is in the white list
func (ip *IP) contains(addr string) (bool, net.IP, error) {
	ipAddr, err := parseIP(addr)
	if err != nil {
		return false, nil, fmt.Errorf("unable to parse address: %s: %s", addr, err)
	}

	contains, err := ip.ContainsIP(ipAddr)
	return contains, ipAddr, err
}

// ContainsIP checks if provided address is in the white list
func (ip *IP) ContainsIP(addr net.IP) (bool, error) {
	if ip.insecure {
		return true, nil
	}

	for _, whiteListIP := range ip.whiteListsIPs {
		if whiteListIP.Equal(addr) {
			return true, nil
		}
	}

	for _, whiteListNet := range ip.whiteListsNet {
		if whiteListNet.Contains(addr) {
			return true, nil
		}
	}

	return false, nil
}

func parseIP(addr string) (net.IP, error) {
	userIP := net.ParseIP(addr)
	if userIP == nil {
		return nil, fmt.Errorf("can't parse IP from address %s", addr)
	}

	return userIP, nil
}

func parseHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
