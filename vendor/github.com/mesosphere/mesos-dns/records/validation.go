package records

import (
	"fmt"
	"net"
)

func validateEnabledServices(c *Config) error {
	if !c.DNSOn && !c.HTTPOn {
		return fmt.Errorf("Either DNS or HTTP server should be on")
	}
	if len(c.Masters) == 0 && c.Zk == "" {
		return fmt.Errorf("specify mesos masters or zookeeper in config.json")
	}
	return nil
}

// validateMasters checks that each master in the list is a properly formatted host:ip pair.
// duplicate masters in the list are not allowed.
// returns nil if the masters list is empty, or else all masters in the list are valid.
func validateMasters(ms []string) error {
	if len(ms) == 0 {
		return nil
	}
	valid := make(map[string]struct{}, len(ms))
	for i, m := range ms {
		h, p, err := net.SplitHostPort(m)
		if err != nil {
			return fmt.Errorf("illegal host:port specified for master %q", ms[i])
		}
		// normalize ipv6 addresses
		if ip := net.ParseIP(h); ip != nil {
			h = ip.String()
			m = h + "_" + p
		}
		//TODO(jdef) distinguish between intended hostnames and invalid ip addresses
		if _, found := valid[m]; found {
			return fmt.Errorf("duplicate master specified: %v", ms[i])
		}
		valid[m] = struct{}{}
	}
	return nil
}

// validateResolvers checks that each resolver in the list is a properly formatted IP address.
// duplicate resolvers in the list are not allowed.
// returns nil if the resolver list is empty, or else all resolvers in the list are valid.
func validateResolvers(rs []string) error {
	if len(rs) == 0 {
		return nil
	}
	ips := make(map[string]struct{}, len(rs))
	for _, r := range rs {
		ip := net.ParseIP(r)
		if ip == nil {
			return fmt.Errorf("illegal IP specified for resolver %q", r)
		}
		ipstr := ip.String()
		if _, found := ips[ipstr]; found {
			return fmt.Errorf("duplicate resolver IP specified: %v", r)
		}
		ips[ipstr] = struct{}{}
	}
	return nil
}

// validateIPSources checks validity of ip sources
func validateIPSources(srcs []string) error {
	if len(srcs) == 0 {
		return fmt.Errorf("empty ip sources")
	}
	if len(srcs) != len(unique(srcs)) {
		return fmt.Errorf("duplicate ip source specified")
	}
	for _, src := range srcs {
		switch src {
		case "host", "docker", "mesos", "netinfo":
		default:
			return fmt.Errorf("invalid ip source %q", src)
		}
	}

	return nil
}
