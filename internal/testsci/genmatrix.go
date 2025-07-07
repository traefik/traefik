package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/packages"
)

const groupCount = 12

type group struct {
	Group string `json:"group"`
}

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedName,
		Dir:  ".",
	}

	pkgs, err := packages.Load(cfg, "./cmd/...", "./pkg/...")
	if err != nil {
		log.Fatal().Err(err).Msg("Loading packages")
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
		log.Fatal().Err(err).Msg("Failed to marshal matrix")
	}

	// Output for GitHub Actions
	fmt.Printf("matrix=%s\n", string(jsonBytes))
}
