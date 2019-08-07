#!/usr/bin/env bash
set -e

for s in apache2 cassandra elasticsearch memcached mysql mongod postgresql sphinxsearch rethinkdb rabbitmq-server redis-server; do sudo service $s stop; done
sudo swapoff -a
sudo dd if=/dev/zero of=/swapfile bs=1M count=3072
sudo mkswap /swapfile
sudo swapon /swapfile
sudo rm -rf /home/runner/.rbenv
#export DOCKER_VERSION=18.06.3
source .semaphoreci/vars
if [ -z "${PULL_REQUEST_NUMBER}" ]; then SHOULD_TEST="-*-"; else TEMP_STORAGE=$(curl --silent https://patch-diff.githubusercontent.com/raw/containous/traefik/pull/${PULL_REQUEST_NUMBER}.diff | patch --dry-run -p1 -R || true); fi
echo ${SHOULD_TEST}
if [ -n "$TEMP_STORAGE" ]; then SHOULD_TEST=$(echo "$TEMP_STORAGE" | grep -Ev '(.md|.yaml|.yml)' || :); fi
echo ${TEMP_STORAGE}
echo ${SHOULD_TEST}
#if [ -n "$SHOULD_TEST" ]; then sudo -E apt-get -yq update; fi
#if [ -n "$SHOULD_TEST" ]; then sudo -E apt-get -yq --no-install-suggests --no-install-recommends --force-yes install docker-ce=${DOCKER_VERSION}*; fi
if [ -n "$SHOULD_TEST" ]; then   docker version; fi
if [ -f "./.semaphoreci/golang.sh" ]; then ./.semaphoreci/golang.sh; fi
if [ -f "./.semaphoreci/golang.sh" ]; then export GOROOT="/usr/local/golang/1.12/go"; fi
if [ -f "./.semaphoreci/golang.sh" ]; then export GOTOOLDIR="/usr/local/golang/1.12/go/pkg/tool/linux_amd64"; fi
