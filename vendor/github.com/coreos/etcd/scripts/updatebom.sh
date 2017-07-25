#!/usr/bin/env bash

set -e

if ! [[ "$0" =~ "scripts/updatebom.sh" ]]; then
	echo "must be run from repository root"
	exit 255
fi

echo "installing 'bill-of-materials.json'"
go get -v -u github.com/coreos/license-bill-of-materials

echo "setting up GOPATH"
rm -rf ./gopath
mkdir ./gopath
mv ./cmd/vendor ./gopath/src

echo "generating bill-of-materials.json"
GOPATH=`pwd`/gopath license-bill-of-materials \
    --override-file ./bill-of-materials.override.json \
    github.com/coreos/etcd github.com/coreos/etcd/etcdctl > bill-of-materials.json

echo "reverting GOPATH,vendor"
mv ./gopath/src ./cmd/vendor
rm -rf ./gopath

echo "generated bill-of-materials.json"

