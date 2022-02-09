#!/usr/bin/env bash
set -e

if [ -z "$VERSION" ]; then
    VERSION=$(git rev-parse HEAD)
fi

if [ -z "$CODENAME" ]; then
    CODENAME=cheddar
fi

if [ -z "$DATE" ]; then
    DATE=$(date -u '+%Y-%m-%d_%I:%M:%S%p')
fi

echo "Building ${VERSION} ${CODENAME} ${DATE}"

GIT_REPO_URL='github.com/traefik/traefik/pkg/version'
GO_BUILD_CMD="go build -ldflags"
GO_BUILD_OPT="-s -w -X ${GIT_REPO_URL}.Version=${VERSION} -X ${GIT_REPO_URL}.Codename=${CODENAME} -X ${GIT_REPO_URL}.BuildDate=${DATE}"

# Build amd64 binaries
OS_PLATFORM_ARG=(linux windows darwin)
OS_ARCH_ARG=(amd64)
for OS in "${OS_PLATFORM_ARG[@]}"; do
  BIN_EXT=''
  if [ "$OS" == "windows" ]; then
    BIN_EXT='.exe'
  fi
  for ARCH in "${OS_ARCH_ARG[@]}"; do
    echo "Building binary for ${OS}/${ARCH}..."
    GOARCH=${ARCH} GOOS=${OS} CGO_ENABLED=0 ${GO_BUILD_CMD} "${GO_BUILD_OPT}" -o "dist/traefik_${OS}-${ARCH}${BIN_EXT}" ./cmd/traefik/
  done
done

# Build arm64 binaries
OS_PLATFORM_ARG=(linux)
OS_ARCH_ARG=(arm64)
for OS in "${OS_PLATFORM_ARG[@]}"; do
  for ARCH in "${OS_ARCH_ARG[@]}"; do
    echo "Building binary for ${OS}/${ARCH}..."
    GOARCH=${ARCH} GOOS=${OS} CGO_ENABLED=0 ${GO_BUILD_CMD} "${GO_BUILD_OPT}" -o "dist/traefik_${OS}-${ARCH}" ./cmd/traefik/
  done
done
