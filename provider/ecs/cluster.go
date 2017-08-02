package ecs

import (
	"fmt"
	"strings"
)

// Clusters holds ecs clusters name
type Clusters []string

//Set adds strings elem into the the parser
//it splits str on , and ;
func (ns *Clusters) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*ns = append(*ns, slice...)
	return nil
}

//Get []string
func (ns *Clusters) Get() interface{} { return Clusters(*ns) }

//String return slice in a string
func (ns *Clusters) String() string { return fmt.Sprintf("%v", *ns) }

//SetValue sets []string into the parser
func (ns *Clusters) SetValue(val interface{}) {
	*ns = Clusters(val.(Clusters))
}
