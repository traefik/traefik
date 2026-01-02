// Command ech provides utilities for generating ECH (Encrypted Client Hello) keys.
//
// Usage:
//
//	go run ./internal/ech/cmd generate example.com,example.org
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/traefik/traefik/v3/internal/ech"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "generate":
		names := strings.Split(os.Args[2], ",")
		if err := ech.GenerateMultiple(os.Stdout, names); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./internal/ech/cmd <command> [arguments]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  generate <sni,sni,...>  Generate ECH keys for the given SNI names (comma-separated)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  go run ./internal/ech/cmd generate example.com")
	fmt.Fprintln(os.Stderr, "  go run ./internal/ech/cmd generate example.com,example.org")
}
