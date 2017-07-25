#!/usr/bin/env bash
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that.
cd $DIR

# Do a hermetic build inside a Docker container.
if [ -z $NOBUILD ]; then
    docker build -t hashicorp/consul-builder scripts/consul-builder/
    docker run --rm -v "$(pwd)":/gopath/src/github.com/hashicorp/consul hashicorp/consul-builder ./scripts/ui_build.sh
fi

exit 0
