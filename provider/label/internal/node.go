package internal

import "reflect"

// Node a label node.
type Node struct {
	Name      string       `json:"name"`
	FieldName string       `json:"fieldName"`
	Value     string       `json:"value,omitempty"`
	Disabled  bool         `json:"disabled,omitempty"`
	Kind      reflect.Kind `json:"kind,omitempty"`
	Children  []*Node      `json:"children,omitempty"`
}
