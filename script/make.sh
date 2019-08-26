#!/usr/bin/env bash
set -e

export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# List of bundles to create when no argument is passed
DEFAULT_BUNDLES=(
	generate
	validate-lint
	binary

	test-unit
	test-integration
)

SCRIPT_DIR="$(cd "$(dirname "${0}")" && pwd -P)"

bundle() {
    local bundle="$1"; shift
    echo "---> Making bundle: $(basename "$bundle") (in $SCRIPT_DIR)"
    # shellcheck source=/dev/null
    source "${SCRIPT_DIR}/$bundle"
}

if [ $# -lt 1 ]; then
    bundles=${DEFAULT_BUNDLES[*]}
else
    bundles=${*}
fi
for bundle in ${bundles[*]}; do
    bundle "$bundle"
    echo
done
