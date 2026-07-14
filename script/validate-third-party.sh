#!/usr/bin/env bash
set -eo pipefail

timestamp_file="$(pwd)/third_party/.last_generated_at"

if [[ ! -f "${timestamp_file}" ]]; then
    echo "Timestamp file not found: ${timestamp_file}"
    echo "Please run 'make third-party' to generate third-party files first."
    exit 1
fi

generated_at="$(tr -d '[:space:]' < "${timestamp_file}")"

if [[ ! "${generated_at}" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
    echo "Invalid timestamp format in ${timestamp_file}: ${generated_at}"
    exit 1
fi

generated_date="${generated_at%%T*}"
release_date="$(date -u '+%Y-%m-%d')"

echo "Generated at: ${generated_date}"
echo "Release date: ${release_date}"

if [[ "${generated_date}" != "${release_date}" ]]; then
    echo "Error: Third-party files were not generated on the same day as the release."
    echo "Please run 'make third-party' to regenerate third-party files for this release."
    exit 1
fi

echo "Third-party files are up-to-date for this release."