#!/bin/bash
# This script will run a couple of linter on the documentation

set -eu

# We want to run all linters before returning success (exit 0) or failure (exit 1)
# So this variable holds the global exit code
EXIT_CODE=0
readonly BASE_DIR="${1:-/app}"

# Run YAML linter for Kubernetes multi-resource files
./docs/scripts/lint-yaml.sh "${BASE_DIR}" || EXIT_CODE=1

echo "== Linting Markdown"
# Uses the file ".markdownlint.json" for setup
cd "${BASE_DIR}" || exit 1

LINTER_EXCLUSIONS="$(find "content" -type f -name '.markdownlint.json')"
GLOBAL_LINT_OPTIONS="--config .markdownlint.json"

# Lint the specific folders (containing linter specific rulesets)
for LINTER_EXCLUSION in ${LINTER_EXCLUSIONS}
do
	markdownlint --config "${LINTER_EXCLUSION}" "$(dirname "${LINTER_EXCLUSION}")" || EXIT_CODE=1
	# Add folder to the ignore list for global lint
	GLOBAL_LINT_OPTIONS="${GLOBAL_LINT_OPTIONS} --ignore=$(dirname "${LINTER_EXCLUSION}")"
done

# Lint all the content, excluding the previously done`
eval markdownlint "${GLOBAL_LINT_OPTIONS}" "content/**/*.md" || EXIT_CODE=1

exit "${EXIT_CODE}"
