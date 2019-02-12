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
markdownlint "${BASE_DIR}/content/**/*.md" || EXIT_CODE=1

echo "== Checking English Writing Style with 'Write-Good'"
write-good "${BASE_DIR}/content/"**/*.md || EXIT_CODE=1

exit "${EXIT_CODE}"
