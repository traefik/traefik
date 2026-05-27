package ip

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/phemmer/go-iptrie"
)

// Checker allows to check that addresses are in a trusted IPs.
type Checker struct {
	trie *iptrie.Trie
}

// NewChecker builds a new Checker given a list of CIDR-Strings to trusted IPs.
func NewChecker(trustedIPs []string) (*Checker, error) {
	if len(trustedIPs) == 0 {
		return nil, errors.New("no trusted IPs provided")
	}

	checker := &Checker{trie: iptrie.NewTrie()}

	for _, ipMask := range trustedIPs {
		if ipAddr := net.ParseIP(ipMask); ipAddr != nil {
			addr, ok := netip.AddrFromSlice(ipAddr)
			if !ok {
				return nil, fmt.Errorf("parsing trusted IPs %s", ipAddr)
			}
			checker.trie.Insert(netip.PrefixFrom(addr, 32*(len(ipAddr)/4)), struct{}{})
		} else {
			_, ipAddr, err := net.ParseCIDR(ipMask)
			if err != nil {
				return nil, fmt.Errorf("parsing CIDR trusted IPs %s: %w", ipAddr, err)
			}
			addr, _ := netip.AddrFromSlice(ipAddr.IP)
			ones, _ := ipAddr.Mask.Size()
			checker.trie.Insert(netip.PrefixFrom(addr, ones), struct{}{})
		}
	}

	return checker, nil
}

// IsAuthorized checks if provided request is authorized by the trusted IPs.
func (ip *Checker) IsAuthorized(addr string) error {
	var invalidMatches []string

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ok, err := ip.Contains(host)
	if err != nil {
		return err
	}

	if !ok {
		invalidMatches = append(invalidMatches, addr)
		return fmt.Errorf("%q matched none of the trusted IPs", strings.Join(invalidMatches, ", "))
	}

	return nil
}

// Contains checks if provided address is in the trusted IPs.
func (ip *Checker) Contains(addr string) (bool, error) {
	if len(addr) == 0 {
		return false, errors.New("empty IP address")
	}

	ipAddr, err := parseIP(addr)
	if err != nil {
		return false, fmt.Errorf("unable to parse address: %s: %w", addr, err)
	}

	return ip.ContainsIP(ipAddr), nil
}

// ContainsIP checks if provided address is in the trusted IPs.
func (ip *Checker) ContainsIP(addr net.IP) bool {
	a, ok := netip.AddrFromSlice(addr)
	if !ok {
		return false
	}
	return ip.trie.Contains(a)
}

func parseIP(addr string) (net.IP, error) {
	parsedAddr, err := netip.ParseAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("can't parse IP from address %s", addr)
	}

	ip := parsedAddr.As16()
	return ip[:], nil
}
