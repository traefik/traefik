#!/usr/bin/env bash
#
# Install dependencies required to cross compile osx, then cleanup
#
# TODO: this should be a separate build stage when CI supports it


set -eu -o pipefail

PKG_DEPS="patch xz-utils clang"

apt-get update -qq
apt-get install -y -q $PKG_DEPS

OSX_SDK=MacOSX10.11.sdk
OSX_CROSS_COMMIT=a9317c18a3a457ca0a657f08cc4d0d43c6cf8953
OSXCROSS_PATH="/osxcross"

echo "Cloning osxcross"
time git clone https://github.com/tpoechtrager/osxcross.git $OSXCROSS_PATH
cd $OSXCROSS_PATH
git checkout -q $OSX_CROSS_COMMIT

echo "Downloading OSX SDK"
time curl -sSL https://s3.dockerproject.org/darwin/v2/${OSX_SDK}.tar.xz \
    -o "${OSXCROSS_PATH}/tarballs/${OSX_SDK}.tar.xz"

echo "Buidling osxcross"
UNATTENDED=yes OSX_VERSION_MIN=10.6 ${OSXCROSS_PATH}/build.sh > /dev/null
