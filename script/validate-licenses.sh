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
# Also ignore attribution generator timestamps
GENERATED_TIMESTAMP_REGEX='[Gg]enerated( at)?:[[:space:]]*[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z'

if ! git diff --exit-code --quiet \
    --ignore-matching-lines="${GENERATED_TIMESTAMP_REGEX}" \
    -- 'third_party/' ':!**/*.json'; then
    echo "Error: license attribution files are out of date."
    echo "Run 'make generate-licenses' and commit the changes."
    echo
    echo "Changed files:"
    git diff \
        --ignore-matching-lines="${GENERATED_TIMESTAMP_REGEX}" \
        --stat \
        -- 'third_party/' ':!**/*.json'
    echo
    echo "diff:"
    git diff \
        --ignore-matching-lines="${GENERATED_TIMESTAMP_REGEX}" \
        -- 'third_party/' ':!**/*.json'
    exit 1
fi

echo "License attribution files are up to date."
