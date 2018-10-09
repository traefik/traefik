package types

import (
	"fmt"
	"strings"
)

// DNSResolvers is a list of DNSes that we will try to resolve the challenged FQDN against
type DNSResolvers []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (r *DNSResolvers) String() string {
	return strings.Join(*r, ",")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (r *DNSResolvers) Set(value string) error {
	entryPoints := strings.Split(value, ",")
	if len(entryPoints) == 0 {
		return fmt.Errorf("wrong DNSResolvers format: %s", value)
	}
	for _, entryPoint := range entryPoints {
		*r = append(*r, entryPoint)
	}
	return nil
}

// Get return the DNSResolvers list
func (r *DNSResolvers) Get() interface{} {
	return *r
}

// SetValue sets the DNSResolvers list
func (r *DNSResolvers) SetValue(val interface{}) {
	*r = val.(DNSResolvers)
}

// Type is type of the struct
func (r *DNSResolvers) Type() string {
	return "dnsresolvers"
}
