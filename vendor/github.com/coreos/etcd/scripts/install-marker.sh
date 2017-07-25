#!/usr/bin/env bash

set -e

ARCH=$1

if [ -z "$1" ]; then
    echo "Usage: ${0} [amd64 or darwin], defaulting to 'amd64'" >> /dev/stderr
    ARCH=amd64
fi

MARKER_URL=https://storage.googleapis.com/etcd/test-binaries/marker-v0.4.0-x86_64-unknown-linux-gnu
if [ ${ARCH} == "darwin" ]; then
    MARKER_URL=https://storage.googleapis.com/etcd/test-binaries/marker-v0.4.0-x86_64-apple-darwin
fi

echo "Installing marker"
curl -L ${MARKER_URL} -o ${GOPATH}/bin/marker
chmod 755 ${GOPATH}/bin/marker

${GOPATH}/bin/marker --version
