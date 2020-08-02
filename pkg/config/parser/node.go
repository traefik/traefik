package parser

import "reflect"

// DefaultRootName is the default name of the root node and the prefix of element name from the resources.
const DefaultRootName = "traefik"

// MapNamePlaceholder is the placeholder for the map name.
const MapNamePlaceholder = "<name>"

// Node is a label node.
type Node struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	FieldName   string            `json:"fieldName"`
	Value       string            `json:"value,omitempty"`
	RawValue    interface{}       `json:"rawValue,omitempty"`
	Disabled    bool              `json:"disabled,omitempty"`
	Kind        reflect.Kind      `json:"kind,omitempty"`
	Tag         reflect.StructTag `json:"tag,omitempty"`
	Children    []*Node           `json:"children,omitempty"`
}
