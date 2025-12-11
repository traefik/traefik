#!/bin/sh
# This script checks that YAML files with multiple Kubernetes resources
# do not start with '---'
#
# Rule: If a YAML file contains more than one Kubernetes resource
# (indicated by '---' separator in the middle of the file),
# it should NOT start with '---'

set -eu

BASE_DIR="${1:-/app}"

echo "== Linting YAML files (Kubernetes multi-resource format)"

# Find all YAML files in the content directory
find "${BASE_DIR}/content" -type f \( -name "*.yml" -o -name "*.yaml" \) | while read -r file; do
    # Count the number of '---' lines in the file
    separator_count=$(grep -c "^---" "$file" || true)

    # Check if file starts with '---'
    starts_with_separator=false
    if head -1 "$file" | grep -q "^---"; then
        starts_with_separator=true
    fi

    # If file has multiple resources (separator_count >= 1 when starting with ---, or >= 2 otherwise)
    # and starts with '---', it's an error
    #
    # Logic:
    # - If starts with '---' and has more than 1 separator -> multiple resources, error
    # - If doesn't start with '---' and has 1+ separators -> multiple resources, ok
    if [ "$starts_with_separator" = true ] && [ "$separator_count" -gt 1 ]; then
        echo "ERROR: $file starts with '---' but contains multiple Kubernetes resources"
        echo "       Files with multiple resources should not start with '---'"
        # We need to signal error but can't use EXIT_CODE in subshell
        # So we output to a temp file
        echo "1" > /tmp/yaml_lint_error
    fi
done

# Check if any errors were found
if [ -f /tmp/yaml_lint_error ]; then
    rm -f /tmp/yaml_lint_error
    exit 1
fi

echo "YAML lint passed"
exit 0
