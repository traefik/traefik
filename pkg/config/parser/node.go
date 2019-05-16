package parser

import "reflect"

// MapNamePlaceholder the placeholder for the map name.
const MapNamePlaceholder = "<name>"

// Node a label node.
type Node struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	FieldName   string            `json:"fieldName"`
	Value       string            `json:"value,omitempty"`
	Disabled    bool              `json:"disabled,omitempty"`
	Kind        reflect.Kind      `json:"kind,omitempty"`
	Tag         reflect.StructTag `json:"tag,omitempty"`
	Children    []*Node           `json:"children,omitempty"`
}
