// +build !go1.6

package render

import (
	"fmt"
	"html/template"
)

// Included helper functions for use when rendering HTML.
var helperFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	// `block` is deprecated! Use the `partial` below if you need this functionality still.
	// Otherwise, checkout Go's `block` implementation introduced in 1.6
	"block": func() (string, error) {
		return "", fmt.Errorf("block called with no layout defined")
	},
	"partial": func() (string, error) {
		return "", fmt.Errorf("block called with no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	},
}
