package types

import (
	"fmt"
	"strings"
)

// Domain holds a domain name with SANs
type Domain struct {
	Main string
	SANs []string
}

// Domains parse []Domain
type Domains []Domain

// Set []Domain
func (ds *Domains) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}

	// get function
	slice := strings.FieldsFunc(str, fargs)
	if len(slice) < 1 {
		return fmt.Errorf("parse error ACME.Domain. Unable to parse %s", str)
	}

	d := Domain{
		Main: slice[0],
	}

	if len(slice) > 1 {
		d.SANs = slice[1:]
	}

	*ds = append(*ds, d)
	return nil
}

// Get []Domain
func (ds *Domains) Get() interface{} { return []Domain(*ds) }

// String returns []Domain in string
func (ds *Domains) String() string { return fmt.Sprintf("%+v", *ds) }

// SetValue sets []Domain into the parser
func (ds *Domains) SetValue(val interface{}) {
	*ds = val.([]Domain)
}
