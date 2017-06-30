#!/usr/bin/env bash
set -e

sudo -E apt-get -yq update
sudo -E apt-get -yq --no-install-suggests --no-install-recommends --force-yes install docker-ce=${DOCKER_VERSION}*
docker version

pip install --user -r requirements.txt

make pull-images
ci_retry make validate
