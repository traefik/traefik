// yarn-lock-to-cdx parses a yarn.lock v1 file and outputs a CycloneDX JSON SBOM.
// It captures ALL packages including platform-specific optional dependencies.
//
// Usage: go run script/yarn-lock-to-cdx.go [-input yarn.lock] [-output sbom.cdx.json]
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BOM struct {
	BOMFormat    string      `json:"bomFormat"`
	SpecVersion  string      `json:"specVersion"`
	SerialNumber string      `json:"serialNumber"`
	Version      int         `json:"version"`
	Metadata     Metadata    `json:"metadata"`
	Components   []Component `json:"components"`
}

type Metadata struct {
	Timestamp string `json:"timestamp"`
}

type Component struct {
	Type    string    `json:"type"`
	Name    string    `json:"name"`
	Version string    `json:"version"`
	PURL    string    `json:"purl"`
	Hashes  []Hash    `json:"hashes,omitempty"`
}

type Hash struct {
	Alg     string `json:"alg"`
	Content string `json:"content"`
}

type yarnPackage struct {
	name      string
	version   string
	integrity string
}

func main() {
	input := flag.String("input", "yarn.lock", "Path to yarn.lock file")
	output := flag.String("output", "-", "Output file (- for stdout)")
	flag.Parse()

	packages, err := parseYarnLock(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing yarn.lock: %v\n", err)
		os.Exit(1)
	}

	// Deduplicate by name@version.
	seen := make(map[string]bool)
	var components []Component
	for _, pkg := range packages {
		key := pkg.name + "@" + pkg.version
		if seen[key] {
			continue
		}
		seen[key] = true

		c := Component{
			Type:    "library",
			Name:    pkg.name,
			Version: pkg.version,
			PURL:    npmPURL(pkg.name, pkg.version),
		}

		if hash, alg := parseIntegrity(pkg.integrity); hash != "" {
			c.Hashes = []Hash{{Alg: alg, Content: hash}}
		}

		components = append(components, c)
	}

	bom := BOM{
		BOMFormat:    "CycloneDX",
		SpecVersion:  "1.6",
		SerialNumber: "urn:uuid:" + uuid.New().String(),
		Version:      1,
		Metadata: Metadata{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		Components: components,
	}

	data, err := json.MarshalIndent(bom, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if *output == "-" {
		os.Stdout.Write(data)
	} else {
		if err := os.WriteFile(*output, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "Generated SBOM with %d components from %s\n", len(components), *input)
}

func parseYarnLock(path string) ([]yarnPackage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var packages []yarnPackage
	var current *yarnPackage

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines.
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			if current != nil {
				packages = append(packages, *current)
				current = nil
			}
			continue
		}

		// Package header: "name@version-range", "name@range1", "name@range2":
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if current != nil {
				packages = append(packages, *current)
			}
			name := extractPackageName(line)
			current = &yarnPackage{name: name}
			continue
		}

		if current == nil {
			continue
		}

		trimmed := strings.TrimSpace(line)

		if v, ok := strings.CutPrefix(trimmed, "version "); ok {
			current.version = unquote(v)
		} else if v, ok := strings.CutPrefix(trimmed, "integrity "); ok {
			current.integrity = unquote(v)
		}
	}

	if current != nil {
		packages = append(packages, *current)
	}

	return packages, scanner.Err()
}

// extractPackageName parses the package name from a yarn.lock header line.
// Examples:
//
//	"@babel/core@^7.12.0", "@babel/core@^7.23.9":  →  @babel/core
//	"lodash@^4.17.21":                               →  lodash
func extractPackageName(line string) string {
	// Remove trailing colon.
	line = strings.TrimSuffix(strings.TrimSpace(line), ":")

	// Take the first entry (before any comma).
	first := strings.SplitN(line, ",", 2)[0]
	first = strings.TrimSpace(first)

	// Remove surrounding quotes.
	first = unquote(first)

	// Find the last @ that separates name from version range.
	// For scoped packages like @babel/core@^7.12.0, we need the last @.
	idx := strings.LastIndex(first, "@")
	if idx <= 0 {
		return first
	}

	return first[:idx]
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func npmPURL(name, version string) string {
	// PURL spec: pkg:npm/[@scope/]name@version
	// Scoped packages: @scope/name → namespace=scope, name=name
	if strings.HasPrefix(name, "@") {
		parts := strings.SplitN(name[1:], "/", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("pkg:npm/%s/%s@%s",
				url.PathEscape("@"+parts[0]),
				url.PathEscape(parts[1]),
				version)
		}
	}
	return fmt.Sprintf("pkg:npm/%s@%s", url.PathEscape(name), version)
}

func parseIntegrity(integrity string) (string, string) {
	if integrity == "" {
		return "", ""
	}
	// Format: "sha512-base64hash=="
	parts := strings.SplitN(integrity, "-", 2)
	if len(parts) != 2 {
		return "", ""
	}

	algMap := map[string]string{
		"sha1":   "SHA-1",
		"sha256": "SHA-256",
		"sha512": "SHA-512",
	}

	alg, ok := algMap[parts[0]]
	if !ok {
		return "", ""
	}

	return parts[1], alg
}
