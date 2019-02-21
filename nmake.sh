#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
sudo chown -R willard:willard ${DIR}/static

tag=gitlab-registry.nordstrom.com/gtm/linkerd-sandbox/traefik:v1.7$2

debug() {
    pushd ${DIR}
    docker build -t ${tag} -f Dockerfile.debug .
    docker push ${tag}
    popd
}

release() {
    pushd ${DIR}
    make image
    docker tag containous/traefik ${tag}
    docker push ${tag}
    popd
}

case $1 in
    "debug")
    debug
    ;;
    "release")
    release
    ;;
    *)
    echo "release|debug"
    exit 1
    ;;
esac
