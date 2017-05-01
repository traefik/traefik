#!/usr/bin/env bash
set -e

if [ -n "$TRAVIS_TAG" ] && [ "$DOCKER_VERSION" = "1.10.3" ]; then
  echo "Deploying..."
else
  echo "Skipping deploy"
  exit 0
fi

git config --global user.email "emile@vauge.com"
git config --global user.name "Emile Vauge"

# load ssh key
echo "Loading key..."
openssl aes-256-cbc -d -k "$pass" -in .travis/traefik.id_rsa.enc -out ~/.ssh/traefik.id_rsa
eval "$(ssh-agent -s)"
chmod 600 ~/.ssh/traefik.id_rsa
ssh-add ~/.ssh/traefik.id_rsa

# update traefik-library-image repo (official Docker image)
echo "Updating traefik-library-imag repo..."
git clone git@github.com:containous/traefik-library-image.git
cd traefik-library-image
./update.sh $VERSION
git add -A
echo $VERSION | git commit --file -
echo $VERSION | git tag -a $VERSION --file -
git push -q --follow-tags -u origin master > /dev/null 2>&1

# create docker image emilevauge/traefik (compatibility)
echo "Updating docker emilevauge/traefik image..."
docker login -e $DOCKER_EMAIL -u $DOCKER_USER -p $DOCKER_PASS
docker tag containous/traefik emilevauge/traefik:latest
docker push emilevauge/traefik:latest
docker tag emilevauge/traefik:latest emilevauge/traefik:${VERSION}
docker push emilevauge/traefik:${VERSION}

cd ..
rm -Rf traefik-library-image/

echo "Deployed"
