package ip

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"
)

// Checker allows to check that addresses are in a trusted IPs.
type Checker struct {
	authorizedIPs    map[netip.Addr]struct{}
	authorizedRanges []ipRange
}

// ipRange represents a contiguous range of IP addresses.
type ipRange struct {
	start netip.Addr
	end   netip.Addr
}

// NewChecker builds a new Checker given a list of CIDR-Strings to trusted IPs.
func NewChecker(trustedIPs []string) (*Checker, error) {
	if len(trustedIPs) == 0 {
		return nil, errors.New("no trusted IPs provided")
	}

	checker := &Checker{
		authorizedIPs: make(map[netip.Addr]struct{}),
	}

	for _, ipMask := range trustedIPs {
		if addr, err := netip.ParseAddr(ipMask); err == nil {
			checker.authorizedIPs[addr.Unmap()] = struct{}{}
			continue
		}

		prefix, err := netip.ParsePrefix(ipMask)
		if err != nil {
			return nil, fmt.Errorf("parsing CIDR trusted IPs %s: %w", ipMask, err)
		}

		startAddr := prefix.Masked().Addr().Unmap()

		if prefix.IsSingleIP() {
			checker.authorizedIPs[startAddr] = struct{}{}
			continue
		}

		endAddr := networkLastAddress(prefix).Unmap()

		checker.authorizedRanges = append(checker.authorizedRanges, ipRange{
			start: startAddr,
			end:   endAddr,
		})
	}

	// Sort and merge overlapping ranges for efficient binary search
	checker.mergeRanges()

	return checker, nil
}

// IsAuthorized checks if provided request is authorized by the trusted IPs.
func (ip *Checker) IsAuthorized(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ok, err := ip.Contains(host)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("%q matched none of the trusted IPs", addr)
	}

	return nil
}

// Contains checks if provided address is in the trusted IPs.
func (ip *Checker) Contains(addr string) (bool, error) {
	if len(addr) == 0 {
		return false, errors.New("empty IP address")
	}

	ipAddr, err := netip.ParseAddr(addr)
	if err != nil {
		return false, fmt.Errorf("unable to parse IP from address %s", addr)
	}

	return ip.containsAddr(ipAddr), nil
}

// ContainsIP checks if provided address is in the trusted IPs.
func (ip *Checker) ContainsIP(addr net.IP) bool {
	ipAddr, ok := netip.AddrFromSlice(addr)
	if !ok {
		return false
	}
	return ip.containsAddr(ipAddr)
}

// containsAddr checks if the provided netip.Addr is in the trusted IPs.
func (ip *Checker) containsAddr(addr netip.Addr) bool {
	addr = addr.Unmap()

	// Check single IPs first (O(1) lookup)
	if _, exists := ip.authorizedIPs[addr]; exists {
		return true
	}

	// Binary search in sorted ranges (O(log n) lookup)
	return ip.containsInRange(addr)
}

// containsInRange uses binary search to check if addr falls within any authorized range.
func (ip *Checker) containsInRange(addr netip.Addr) bool {
	ranges := ip.authorizedRanges
	if len(ranges) == 0 {
		return false
	}

	idx, _ := slices.BinarySearchFunc(ranges, addr, func(r ipRange, target netip.Addr) int {
		return r.end.Compare(target)
	})

	return idx < len(ranges) && ranges[idx].start.Compare(addr) <= 0
}

// mergeRanges sorts and merges overlapping/adjacent IP ranges.
func (ip *Checker) mergeRanges() {
	ranges := ip.authorizedRanges
	if len(ranges) <= 1 {
		return
	}

	// Sort ranges by start address
	slices.SortFunc(ranges, func(a, b ipRange) int {
		return a.start.Compare(b.start)
	})

	// Merge overlapping/adjacent ranges
	merged := make([]ipRange, 0, len(ranges))
	current := ranges[0]

	for _, next := range ranges[1:] {
		// Check if ranges overlap or are adjacent.
		// If current.end is the max IP, Next() overflows to zero, which is always < next.start.
		nextEnd := current.end.Next()
		if !nextEnd.IsValid() || nextEnd.Compare(next.start) >= 0 {
			// Merge: extend current range if next.end is larger
			if next.end.Compare(current.end) > 0 {
				current.end = next.end
			}
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)

	ip.authorizedRanges = merged
}

// networkLastAddress calculates the last IP address in a CIDR block.
func networkLastAddress(prefix netip.Prefix) netip.Addr {
	addr := prefix.Addr()
	bits := prefix.Bits()

	ipNet := &net.IPNet{
		IP:   addr.AsSlice(),
		Mask: net.CIDRMask(bits, addr.BitLen()),
	}

	last := make(net.IP, len(ipNet.IP))
	for i := range ipNet.IP {
		last[i] = ipNet.IP[i] | ^ipNet.Mask[i]
	}

	res, _ := netip.AddrFromSlice(last)
	return res
}
