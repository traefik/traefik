#!/usr/bin/env bash
set -o errexit
set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${0}" )" && pwd -P)"; export SCRIPT_DIR

echo "checking go modules for unintentional changes..."

go mod tidy
git diff --exit-code go.mod
git diff --exit-code go.sum

echo 'Congratulations! All go modules changes are done the right way.'
