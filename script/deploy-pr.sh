#!/bin/bash
set -e

if ([ "$TRAVIS_BRANCH" = "master" ] && [ -z "$TRAVIS_TAG" ]) && [ "$TRAVIS_PULL_REQUEST" = "false" ] && [ "$DOCKER_VERSION" = "1.10.1" ]; then
  echo "Deploying PR..."
else
  echo "Skipping deploy PR"
  exit 0
fi

COMMENT=$(git log -1 --pretty=%B)
PR=$(echo $COMMENT | grep -oP "Merge pull request #\K(([0-9]*))(?=.*)")

if [ -z "$PR" ]; then
    echo "Unable to get PR number: $PR from: $COMMENT"
    exit 0
fi  

# create docker image containous/traefik
echo "Updating docker containous/traefik image..."
docker login -e $DOCKER_EMAIL -u $DOCKER_USER -p $DOCKER_PASS
docker tag containous/traefik containous/traefik:pr-${PR}
docker push containous/traefik:pr-${PR}
docker tag containous/traefik containous/traefik:experimental
docker push containous/traefik:experimental

echo "Deployed"
