#!/bin/bash -eu

git describe --tags "$VERSION" > /dev/null || exit 1

go get github.com/mitchellh/gox

gox -arch=amd64 \
    -os="linux darwin windows" \
    -output="{{.Dir}}-${VERSION}-{{.OS}}-{{.Arch}}" \
    -ldflags="-X main.Version=${VERSION}"

gzip --best mesos-dns-${VERSION}-*
