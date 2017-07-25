#!/usr/bin/env bash
#
# Build all release binaries and images to directory ./release.
# Run from repository root.
#
set -e

VERSION=$1
if [ -z "${VERSION}" ]; then
	echo "Usage: ${0} VERSION" >> /dev/stderr
	exit 255
fi

if ! command -v acbuild >/dev/null; then
    echo "cannot find acbuild"
    exit 1
fi

if ! command -v docker >/dev/null; then
    echo "cannot find docker"
    exit 1
fi

ETCD_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

pushd ${ETCD_ROOT} >/dev/null
	echo Building etcd binary...
	./scripts/build-binary ${VERSION}

	# ppc64le not yet supported by acbuild.
	for TARGET_ARCH in "amd64" "arm64"; do
		echo Building ${TARGET_ARCH} aci image...
		GOARCH=${TARGET_ARCH} BINARYDIR=release/etcd-${VERSION}-linux-${TARGET_ARCH} BUILDDIR=release ./scripts/build-aci ${VERSION}
	done

	for TARGET_ARCH in "amd64" "arm64" "ppc64le"; do
		echo Building ${TARGET_ARCH} docker image...
		GOARCH=${TARGET_ARCH} BINARYDIR=release/etcd-${VERSION}-linux-${TARGET_ARCH} BUILDDIR=release ./scripts/build-docker ${VERSION}
	done
popd >/dev/null
