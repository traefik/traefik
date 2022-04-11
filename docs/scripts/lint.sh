#!/bin/sh
# This script will run a couple of linter on the documentation

set -eu

# We want to run all linters before returning success (exit 0) or failure (exit 1)
# So this variable holds the global exit code
EXIT_CODE=0
readonly BASE_DIR=/app

echo "== Linting Markdown"
# Uses the file ".markdownlint.json" for setup
cd "${BASE_DIR}" || exit 1

LINTER_EXCLUSIONS="$(find "${BASE_DIR}/content" -type f -name '.markdownlint.json')"
GLOBAL_LINT_OPTIONS="--config ${BASE_DIR}/.markdownlint.json"

# Lint the specific folders (containing linter specific rulesets)
for LINTER_EXCLUSION in ${LINTER_EXCLUSIONS}
do
	markdownlint --config "${LINTER_EXCLUSION}" "$(dirname "${LINTER_EXCLUSION}")" || EXIT_CODE=1
	# Add folder to the ignore list for global lint
	GLOBAL_LINT_OPTIONS="${GLOBAL_LINT_OPTIONS} --ignore=$(dirname "${LINTER_EXCLUSION}")"
done

# Lint all the content, excluding the previously done`
eval markdownlint "${GLOBAL_LINT_OPTIONS}" "${BASE_DIR}/content/**/*.md" || EXIT_CODE=1

exit "${EXIT_CODE}"
