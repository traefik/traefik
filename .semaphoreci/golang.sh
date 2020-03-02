#!/usr/bin/env bash

set -e

curl -O https://dl.google.com/go/go1.14.linux-amd64.tar.gz

tar -xvf go1.14.linux-amd64.tar.gz
rm -rf go1.14.linux-amd64.tar.gz

sudo mkdir -p /usr/local/golang/1.14/go
sudo mv go /usr/local/golang/1.14/

sudo rm /usr/local/bin/go
sudo chmod +x /usr/local/golang/1.14/go/bin/go
sudo ln -s /usr/local/golang/1.14/go/bin/go /usr/local/bin/go

export GOROOT="/usr/local/golang/1.14/go"
export GOTOOLDIR="/usr/local/golang/1.14/go/pkg/tool/linux_amd64"

go version
