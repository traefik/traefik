#!/bin/bash
set -e

# Get the version from the command line.
VERSION=$1
if [ -z $VERSION ]; then
    echo "Please specify a version."
    exit 1
fi

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that.
cd $DIR

# Generate the tag.
if [ -z $NOTAG ]; then
  echo "==> Tagging..."
  git commit --allow-empty -a --gpg-sign=348FFC4C -m "Release v$VERSION"
  git tag -a -m "Version $VERSION" -s -u 348FFC4C "v${VERSION}" master
fi

# Do a hermetic build inside a Docker container.
if [ -z $NOBUILD ]; then
    docker build -t hashicorp/serf-builder scripts/serf-builder/
    docker run --rm -v "$(pwd)":/gopath/src/github.com/hashicorp/serf hashicorp/serf-builder
fi

# Zip all the files.
rm -rf ./pkg/dist
mkdir -p ./pkg/dist
for FILENAME in $(find ./pkg -mindepth 1 -maxdepth 1 -type f); do
  FILENAME=$(basename $FILENAME)
  cp ./pkg/${FILENAME} ./pkg/dist/serf_${VERSION}_${FILENAME}
done

# Make the checksums.
pushd ./pkg/dist
shasum -a256 * > ./serf_${VERSION}_SHA256SUMS
if [ -z $NOSIGN ]; then
  echo "==> Signing..."
  gpg --default-key 348FFC4C --detach-sig ./serf_${VERSION}_SHA256SUMS
fi
popd

exit 0
