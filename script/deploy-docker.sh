#!/usr/bin/env bash
set -e

if [ -n "$TRAVIS_COMMIT" ]; then
  echo "Deploying PR..."
else
  echo "Skipping deploy PR"
  exit 0
fi

# create docker image joshdvir/traefik
echo "Updating docker joshdvir/traefik image..."
docker login -u $DOCKER_USER -p $DOCKER_PASS
docker tag joshdvir/traefik joshdvir/traefik:${TRAVIS_COMMIT}
docker push joshdvir/traefik:${TRAVIS_COMMIT}
docker tag joshdvir/traefik joshdvir/traefik:experimental
docker push joshdvir/traefik:experimental

echo "Deployed"
