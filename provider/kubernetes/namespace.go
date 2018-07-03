package kubernetes

import (
	"fmt"
	"strings"
)

// Namespaces holds kubernetes namespaces
type Namespaces []string

// Set adds strings elem into the the parser
// it splits str on , and ;
func (ns *Namespaces) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*ns = append(*ns, slice...)
	return nil
}

// Get []string
func (ns *Namespaces) Get() interface{} { return *ns }

// String return slice in a string
func (ns *Namespaces) String() string { return fmt.Sprintf("%v", *ns) }

// SetValue sets []string into the parser
func (ns *Namespaces) SetValue(val interface{}) {
	*ns = val.(Namespaces)
}
