package kubernetes

import (
       "fmt"
)

// IngressClasses holds kubernetes ingressClass annotation values
type IngressClasses []string

//Set adds strings elem into the the parser
//it splits str on , and ;
func (ic *IngressClasses) Set(str string) error {
       *ic = append(*ic, str)
       return nil
}

//Get []string
func (ic *IngressClasses) Get() interface{} { return *ic }

//String return slice in a string
func (ic *IngressClasses) String() string { return fmt.Sprintf("%v", *ic) }

//SetValue sets []string into the parser
func (ic *IngressClasses) SetValue(val interface{}) {
       *ic = val.(IngressClasses)
}
