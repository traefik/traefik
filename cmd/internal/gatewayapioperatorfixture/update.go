package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// refreshUpstream overwrites ./upstream with a fresh copy of the manifests,
// read from a local checkout of source, or fetched at the source ref through
// the GitHub CLI (so that access to the private repository is the caller's,
// not this tool's, to manage).
func refreshUpstream(source string) error {
	local, err := isDir(source)
	if err != nil {
		return err
	}

	if !local {
		if _, err := exec.LookPath("gh"); err != nil {
			return fmt.Errorf("the GitHub CLI is required to read %s", repository)
		}
	}

	for _, manifest := range manifests {
		content, err := fetch(source, local, manifest)
		if err != nil {
			return err
		}

		dest := filepath.Join(upstream, manifest)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(dest, content, 0o644); err != nil {
			return err
		}
	}

	rev, err := resolveRevision(source, local)
	if err != nil {
		return err
	}

	return writeProvenance(rev)
}

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return info.IsDir(), nil
}

// fetch reads a manifest either from a local checkout or, with retries, from
// the repository at the given ref.
func fetch(source string, local bool, manifest string) ([]byte, error) {
	if local {
		path := filepath.Join(source, manifest)

		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("missing %s: is %s a %s checkout?", path, source, repository)
		}

		return content, nil
	}

	const attempts = 3

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		content, err := fetchRemote(source, manifest)
		if err == nil {
			return content, nil
		}

		lastErr = err
		fmt.Fprintf(os.Stderr, "Retrying %s (attempt %d): %v\n", manifest, attempt, err)
		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("unable to read %s from %s@%s: %w", manifest, repository, source, lastErr)
}

func fetchRemote(source, manifest string) ([]byte, error) {
	out, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/contents/%s?ref=%s", repository, manifest, source),
		"--jq", ".content").Output()
	if err != nil {
		return nil, err
	}

	// The GitHub API wraps the base64 content across multiple lines.
	clean := strings.Join(strings.Fields(string(out)), "")

	return base64.StdEncoding.DecodeString(clean)
}

// resolveRevision reports the exact commit ./upstream was vendored from, so a
// later refresh can be diffed against a known point instead of a moving ref.
func resolveRevision(source string, local bool) (string, error) {
	if local {
		out, err := exec.Command("git", "-C", source, "rev-parse", "HEAD").Output()
		if err != nil {
			return "", fmt.Errorf("resolving %s HEAD: %w", source, err)
		}
		rev := strings.TrimSpace(string(out))

		status, err := exec.Command("git", "-C", source, "status", "--porcelain").Output()
		if err != nil {
			return "", fmt.Errorf("checking %s worktree status: %w", source, err)
		}
		if len(bytes.TrimSpace(status)) > 0 {
			rev += " (with uncommitted local changes)"
		}

		return rev, nil
	}

	out, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/commits/%s", repository, source), "--jq", ".sha").Output()
	if err != nil {
		return "", fmt.Errorf("resolving %s@%s: %w", repository, source, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func writeProvenance(rev string) error {
	content := fmt.Sprintf(`This directory is a point-in-time, unmodified copy of the %[2]sconfig/crd%[2]s,
%[2]sconfig/rbac%[2]s and %[2]sconfig/manager%[2]s directories of
[traefik/gateway-operator](https://github.com/traefik/gateway-operator), a
private repository.

It exists so that %[2]sgo run ./cmd/internal/gatewayapioperatorfixture%[2]s can render
%[2]sintegration/fixtures/gateway-api-conformance/01-operator.yml%[2]s without network
access or access to that repository. Refresh it, deliberately and rarely, with:

%[3]s
go run ./cmd/internal/gatewayapioperatorfixture -update [ref|path to a checkout]
%[3]s

Source: %[1]s@%[4]s
Vendored: %[5]s
`, repository, "`", "```", rev, time.Now().Format("2006-01-02"))

	return os.WriteFile(filepath.Join(upstream, "PROVENANCE.md"), []byte(content), 0o644)
}
