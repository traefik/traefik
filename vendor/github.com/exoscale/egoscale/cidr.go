package egoscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

// CIDR represents a nicely JSON serializable net.IPNet
type CIDR struct {
	net.IPNet
}

// UnmarshalJSON unmarshals the raw JSON into the MAC address
func (cidr *CIDR) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	c, err := ParseCIDR(s)
	if err != nil {
		return err
	}
	*cidr = CIDR{c.IPNet}

	return nil
}

// MarshalJSON converts the CIDR to a string representation
func (cidr CIDR) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", cidr)), nil
}

// String returns the string representation of a CIDR
func (cidr CIDR) String() string {
	return cidr.IPNet.String()
}

// ParseCIDR parses a CIDR from a string
func ParseCIDR(s string) (*CIDR, error) {
	_, net, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	return &CIDR{*net}, nil
}

// MustParseCIDR forces parseCIDR or panics
func MustParseCIDR(s string) *CIDR {
	cidr, err := ParseCIDR(s)
	if err != nil {
		panic(err)
	}

	return cidr
}

// Equal compare two CIDR
func (cidr CIDR) Equal(c CIDR) bool {
	return (cidr.IPNet.IP.Equal(c.IPNet.IP) && bytes.Equal(cidr.IPNet.Mask, c.IPNet.Mask))
}
