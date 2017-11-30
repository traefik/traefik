package whitelist

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// IP allows to check that addresses are in a white list
type IP struct {
	whiteListsIPs []*net.IP
	whiteListsNet []*net.IPNet
	insecure      bool
}

// NewIP builds a new IP given a list of CIDR-Strings to whitelist
func NewIP(whitelistStrings []string, insecure bool) (*IP, error) {
	if len(whitelistStrings) == 0 && !insecure {
		return nil, errors.New("no whiteListsNet provided")
	}

	ip := IP{}

	if !insecure {
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
	}

	return &ip, nil
}

// Contains checks if provided address is in the white list
func (ip *IP) Contains(addr string) (bool, net.IP, error) {
	if ip.insecure {
		return true, nil, nil
	}

	ipAddr, err := ipFromRemoteAddr(addr)
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

func ipFromRemoteAddr(addr string) (net.IP, error) {
	userIP := net.ParseIP(addr)
	if userIP == nil {
		return nil, fmt.Errorf("can't parse IP from address %s", addr)
	}

	return userIP, nil
}

func GetRemoteIp(req *http.Request, trustProxy *IP) (net.IP, error) {
	remoteIp, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return net.IP{}, err
	}

	ip := net.ParseIP(remoteIp)
	if ip == nil {
		return nil, fmt.Errorf("can't parse IP from address %s", remoteIp)
	}

	// if we trust the upstream host, we can filter based on the
	// client ip it reports
	if trustProxy != nil {
		if contains, _ := trustProxy.ContainsIP(ip); contains {
			if remoteIp := req.Header.Get("X-Forwarded-For"); remoteIp != "" {
				ips := strings.Split(remoteIp, ",")
				for i := len(ips) - 1; i >= 0; i-- {
					if ip := net.ParseIP(strings.TrimSpace(ips[i])); ip != nil {
						// if we trust this host, and there are more upstream hosts
						// then we can report the upstream host
						if contains, _ := trustProxy.ContainsIP(ip); contains && i > 0 {
							continue
						}
						return ip, nil
					}
				}
			}
		}
	}

	return ip, nil
}
