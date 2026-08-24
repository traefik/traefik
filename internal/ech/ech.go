// Package ech provides utilities for generating Encrypted Client Hello keys.
package ech

import (
	"fmt"
	"io"

	"github.com/traefik/traefik/v3/pkg/tls"
)

// Generate creates a new ECH key for the given public name and writes it as PEM.
func Generate(w io.Writer, publicName string) error {
	key, err := tls.NewECHKey(publicName)
	if err != nil {
		return fmt.Errorf("failed to generate ECH key for %s: %w", publicName, err)
	}

	data, err := tls.MarshalECHKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal ECH key for %s: %w", publicName, err)
	}

	if _, err = w.Write(data); err != nil {
		return fmt.Errorf("failed to write ECH key for %s: %w", publicName, err)
	}

	return nil
}
