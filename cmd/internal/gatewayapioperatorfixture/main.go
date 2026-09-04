// Command gatewayapioperatorfixture renders the Traefik Gateway API operator
// install manifest used by the Gateway API conformance suite, from a vendored,
// point-in-time copy of the github.com/traefik/gateway-operator repository (a
// private repository) kept under ./upstream.
//
// Usage:
//
//	go run ./cmd/internal/gatewayapioperatorfixture
//	go run ./cmd/internal/gatewayapioperatorfixture -update [ref|path to a checkout]
//
// The default form only reads ./upstream: no network access, and no access to
// the operator repository, is needed to render the fixture. -update refreshes
// ./upstream first, from a local checkout (renders operator changes that are
// not pushed yet) or a ref (the default, "main", is what keeps a refresh
// reproducible), and then renders.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	repository = "traefik/gateway-operator"
	upstream   = "cmd/internal/gatewayapioperatorfixture/upstream"
	fixture    = "integration/fixtures/gateway-api-conformance/01-operator.yml"
)

// manifests are read from ./upstream, in order, and concatenated into the
// fixture.
var manifests = []string{
	"config/crd/gateway.traefik.io_traefikproxies.yaml",
	"config/rbac/role.yaml",
	"config/rbac/dataplane_role.yaml",
	"config/manager/manager.yaml",
}

func main() {
	update := flag.Bool("update", false, "refresh ./upstream from a ref or a local checkout (the first non-flag argument) before rendering")
	flag.Parse()

	source := "main"
	if args := flag.Args(); len(args) > 0 && args[0] != "" {
		source = args[0]
	}

	if *update {
		if err := refreshUpstream(source); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err := render(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func render() error {
	contents := make(map[string][]byte, len(manifests))

	for _, manifest := range manifests {
		content, err := os.ReadFile(filepath.Join(upstream, manifest))
		if err != nil {
			return fmt.Errorf("reading vendored %s: %w (run with -update first?)", manifest, err)
		}

		// The concatenation below separates the manifests, so a leading
		// separator would open an empty document.
		contents[manifest] = stripLeadingSeparator(content)
	}

	manager, err := pinOperatorImage(contents["config/manager/manager.yaml"])
	if err != nil {
		return err
	}
	contents["config/manager/manager.yaml"] = manager

	var buf bytes.Buffer
	buf.WriteString("# Traefik Gateway API operator install.\n#\n")
	buf.WriteString("# Rendered from a vendored copy of " + repository + " (" + upstream + "); see there for its provenance.\n")
	buf.WriteString("# Regenerate with: make generate-gateway-api-operator-fixture\n")

	for _, manifest := range manifests {
		buf.WriteString("---\n")
		buf.Write(contents[manifest])
	}

	if err := os.WriteFile(fixture, buf.Bytes(), 0o644); err != nil {
		return err
	}

	fmt.Println("Rendered " + fixture)

	return nil
}

func stripLeadingSeparator(content []byte) []byte {
	rest, ok := bytes.CutPrefix(content, []byte("---\n"))
	if !ok {
		return content
	}

	return rest
}

var (
	imageLine           = regexp.MustCompile(`(?m)^( *)image: traefik/gateway-operator:latest$`)
	leaderElectLine     = regexp.MustCompile(`(?m)^ *- --leader-elect\n`)
	imagePullPolicyLine = regexp.MustCompile(`(?m)^ *imagePullPolicy: Never$`)
)

// pinOperatorImage patches the operator manager manifest so the image side
// loaded into the k3s node for the suite is used as is: pulling it would
// resolve the tag to the published image instead. Leader election only delays
// the startup of the single replica.
func pinOperatorImage(manager []byte) ([]byte, error) {
	patched := imageLine.ReplaceAll(manager, []byte("${1}image: traefik/gateway-operator:latest\n${1}imagePullPolicy: Never"))
	patched = leaderElectLine.ReplaceAll(patched, nil)

	if !imagePullPolicyLine.Match(patched) {
		return nil, errors.New("unexpected operator manager manifest: no traefik/gateway-operator:latest image to pin")
	}

	if bytes.Contains(patched, []byte("--leader-elect")) {
		return nil, errors.New("unexpected operator manager manifest: leader election is still enabled")
	}

	return patched, nil
}
