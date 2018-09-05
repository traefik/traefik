#!/usr/bin/env bash
set -e

export DOCKER_VERSION=17.03.1

source .semaphoreci/vars

if [ -z "${PULL_REQUEST_NUMBER}" ]; then SHOULD_TEST="-*-"; else TEMP_STORAGE=$(curl --silent https://patch-diff.githubusercontent.com/raw/containous/traefik/pull/${PULL_REQUEST_NUMBER}.diff | patch --dry-run -p1 -R); fi

if [ -n "$TEMP_STORAGE" ]; then SHOULD_TEST=$(echo "$TEMP_STORAGE" | grep -Ev '(.md|.yaml|.yml)' || :); fi

if [ -n "$SHOULD_TEST" ]; then sudo -E apt-get -yq update; fi

if [ -n "$SHOULD_TEST" ]; then sudo -E apt-get -yq --no-install-suggests --no-install-recommends --force-yes install docker-ce=${DOCKER_VERSION}*; fi

if [ -n "$SHOULD_TEST" ]; then  docker version; fi
