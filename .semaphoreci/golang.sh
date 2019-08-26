#!/usr/bin/env bash

set -e

curl -O https://dl.google.com/go/go"${GO_VERSION}".linux-amd64.tar.gz

tar -xvf go"${GO_VERSION}".linux-amd64.tar.gz
rm -rf go"${GO_VERSION}".linux-amd64.tar.gz

sudo mkdir -p /usr/local/golang/"${GO_VERSION}"/go
sudo mv go /usr/local/golang/"${GO_VERSION}"/

sudo rm /usr/local/bin/go
sudo chmod +x /usr/local/golang/"${GO_VERSION}"/go/bin/go
sudo ln -s /usr/local/golang/"${GO_VERSION}"/go/bin/go /usr/local/bin/go

export GOROOT="/usr/local/golang/${GO_VERSION}/go"
export GOTOOLDIR="/usr/local/golang/${GO_VERSION}/go/pkg/tool/linux_amd64"

go version
