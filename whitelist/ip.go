package whitelist

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
)

// IP allows to check that addresses are in a white list
type IP struct {
	whiteListsIPs []*net.IP
	whiteListsNet []*net.IPNet
}

// NewIP builds a new IP given a list of CIDR-Strings to whitelist
func NewIP(whitelistStrings []string) (*IP, error) {
	if len(whitelistStrings) == 0 {
		return nil, errors.New("no whiteListsNet provided")
	}

	ip := IP{}

	for _, whitelistString := range whitelistStrings {
		ipAddr := net.ParseIP(whitelistString)
		if ipAddr != nil {
			ip.whiteListsIPs = append(ip.whiteListsIPs, &ipAddr)
		} else {
			_, whitelist, err := net.ParseCIDR(whitelistString)
			if err != nil {
				return nil, fmt.Errorf("parsing CIDR whitelist %s: %v", whitelist, err)
			}
			ip.whiteListsNet = append(ip.whiteListsNet, whitelist)
		}
	}

	return &ip, nil
}

// Contains checks if provided address is in the white list
func (ip *IP) Contains(addr string) (bool, net.IP, error) {
	ipAddr, err := ipFromRemoteAddr(addr)
	if err != nil {
		return false, nil, fmt.Errorf("unable to parse address: %s: %s", addr, err)
	}

	contains, err := ip.ContainsIP(ipAddr)
	return contains, ipAddr, err
}

// ContainsIP checks if provided address is in the white list
func (ip *IP) ContainsIP(addr net.IP) (bool, error) {
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

func ipFromRemoteAddr(addr string) (net.IP, error) {
	userIP := net.ParseIP(addr)
	if userIP == nil {
		return nil, fmt.Errorf("can't parse IP from address %s", addr)
	}

	return userIP, nil
}
