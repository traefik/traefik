#!/usr/bin/env bash
set -e

if [ -n "$TRAVIS_COMMIT" ]; then
  echo "Deploying PR..."
else
  echo "Skipping deploy PR"
  exit 0
fi

# create docker image containous/traefik
echo "Updating docker containous/traefik image..."
docker login -u $DOCKER_USER -p $DOCKER_PASS
docker tag containous/traefik containous/traefik:${TRAVIS_COMMIT}
docker push containous/traefik:${TRAVIS_COMMIT}
docker tag containous/traefik containous/traefik:experimental
docker push containous/traefik:experimental

echo "Deployed"
