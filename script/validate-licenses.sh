#!/usr/bin/env bash
set -euo pipefail

# Validates that license attribution files are up to date with dependencies.
# Regenerates and uses git diff to compare with committed files.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_DIR}"

# Regenerate.
"${SCRIPT_DIR}/generate-licenses.sh"

# Compare (ignore sbom/ as it contains timestamps/UUIDs that change every run).
if ! git diff --exit-code --quiet -- 'licenses/' ':!licenses/sbom/'; then
    echo "Error: license attribution files are out of date."
    echo "Run 'make generate-licenses' and commit the changes."
    git diff --stat -- 'licenses/' ':!licenses/sbom/'
    exit 1
fi

echo "License attribution files are up to date."
