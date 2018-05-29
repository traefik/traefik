package ecs

import (
	"fmt"
	"strings"
)

// Clusters holds ecs clusters name
type Clusters []string

// Set adds strings elem into the the parser
// it splits str on , and ;
func (c *Clusters) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*c = append(*c, slice...)
	return nil
}

// Get Clusters
func (c *Clusters) Get() interface{} { return *c }

// String return slice in a string
func (c *Clusters) String() string { return fmt.Sprintf("%v", *c) }

// SetValue sets Clusters into the parser
func (c *Clusters) SetValue(val interface{}) {
	*c = val.(Clusters)
}
