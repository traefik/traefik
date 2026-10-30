// Package ech provides the Encrypted Client Hello commands.
package ech

import (
	"errors"
	"fmt"
	"os"

	"github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v3/pkg/tls"
)

// NewCmd builds a new ECH command.
func NewCmd() (*cli.Command, error) {
	cmd := &cli.Command{
		Name:        "ech",
		Description: `Manages Encrypted Client Hello (ECH) keys.`,
	}

	err := cmd.AddCommand(&cli.Command{
		Name:        "generate",
		Description: `Generates an ECH key for the given public name, written to the standard output in the RFC 9934 PEM format.`,
		AllowArg:    true,
		Run:         generate,
	})
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func generate(args []string) error {
	if len(args) != 1 {
		return errors.New("expected a single public name argument")
	}

	key, err := tls.NewECHKey(args[0])
	if err != nil {
		return fmt.Errorf("generating ECH key: %w", err)
	}

	data, err := tls.MarshalECHKey(key)
	if err != nil {
		return fmt.Errorf("marshaling ECH key: %w", err)
	}

	_, err = os.Stdout.Write(data)
	return err
}
