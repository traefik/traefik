package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/traefik/traefik/v2/pkg/log"
	"golang.org/x/tools/go/packages"
)

const groupCount = 12

type group struct {
	Group string `json:"group"`
}

func main() {
	logger := log.WithoutContext()

	cfg := &packages.Config{
		Mode: packages.NeedName,
		Dir:  ".",
	}

	pkgs, err := packages.Load(cfg, "./cmd/...", "./pkg/...")
	if err != nil {
		logger.Fatalf("Loading packages: %v", err)
	}

	var packageNames []string
	for _, pkg := range pkgs {
		if pkg.PkgPath != "" {
			packageNames = append(packageNames, pkg.PkgPath)
		}
	}

	total := len(packageNames)
	perGroup := (total + groupCount - 1) / groupCount

	fmt.Fprintf(os.Stderr, "Total packages: %d\n", total)
	fmt.Fprintf(os.Stderr, "Packages per group: %d\n", perGroup)

	var matrix []group
	for i := range groupCount {
		start := i * perGroup
		end := start + perGroup
		if start >= total {
			break
		}
		if end > total {
			end = total
		}
		g := strings.Join(packageNames[start:end], " ")
		matrix = append(matrix, group{Group: g})
	}

	jsonBytes, err := json.Marshal(matrix)
	if err != nil {
		logger.Fatalf("Failed to marshal matrix: %v", err)
	}

	// Output for GitHub Actions
	fmt.Printf("matrix=%s\n", string(jsonBytes))
}
