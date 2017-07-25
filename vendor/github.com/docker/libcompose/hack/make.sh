#!/usr/bin/env bash
set -e

export LIBCOMPOSE_PKG='github.com/docker/libcompose'

# List of bundles to create when no argument is passed
DEFAULT_BUNDLES=(
    validate-gofmt
    validate-dco
    validate-git-marks
    validate-lint
    validate-vet
    binary

    test-unit
    test-integration
    test-acceptance

    cross-binary
)
bundle() {
    local bundle="$1"; shift
    echo "---> Making bundle: $(basename "$bundle") (in $DEST)"
    source "hack/$bundle" "$@"
}

if [ $# -lt 1 ]; then
    bundles=(${DEFAULT_BUNDLES[@]})
else
    bundles=($@)
fi
for bundle in ${bundles[@]}; do
    export DEST=.
    ABS_DEST="$(cd "$DEST" && pwd -P)"
    bundle "$bundle"
    echo
done
