#!/bin/sh
set -e

if ([ "$TRAVIS_BRANCH" = "master" ] || [ ! -z "$TRAVIS_TAG" ]) && [ "$TRAVIS_PULL_REQUEST" = "false" ]; then
  echo "Deploying"
else
  echo "Skipping deploy"
  exit 0
fi

curl -LO https://github.com/tcnksm/ghr/releases/download/pre-release/linux_amd64.zip
unzip -q linux_amd64.zip
sudo mv ghr /usr/bin/ghr
sudo chmod +x /usr/bin/ghr

ghr -t $GITHUB_TOKEN -u emilevauge -r traefik --prerelease ${VERSION} dist/
docker login -e $DOCKER_EMAIL -u $DOCKER_USER -p $DOCKER_PASS
docker push ${REPO,,}:latest
docker tag ${REPO,,}:latest ${REPO,,}:${VERSION}
docker push ${REPO,,}:${VERSION}
