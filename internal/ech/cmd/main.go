// Command ech generates Encrypted Client Hello keys.
//
// Usage:
//
//	go run ./internal/ech/cmd generate example.com
package main

import (
	"fmt"
	"os"

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
		if err := ech.Generate(os.Stdout, os.Args[2]); err != nil {
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
	fmt.Fprintln(os.Stderr, "  generate <public-name>  Generate an ECH key for the public name")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  go run ./internal/ech/cmd generate example.com")
}
