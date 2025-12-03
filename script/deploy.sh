#!/usr/bin/env bash
set -e

if [ -n "${VERSION}" ]; then
  echo "Deploying..."
else
  echo "Skipping deploy"
  exit 0
fi

git config --global user.email "${BAQUPER_EMAIL}"
git config --global user.name "Baquper"

# load ssh key
eval "$(ssh-agent -s)"
chmod 600 ~/.ssh/baquper_rsa
ssh-add ~/.ssh/baquper_rsa

# update baqup-library-image repo (official Docker image)
echo "Updating baqup-library-imag repo..."
git clone git@github.com:baqup/baqup-library-image.git
cd baqup-library-image
./updatev2.sh "${VERSION}"
git add -A
echo "${VERSION}" | git commit --file -
echo "${VERSION}" | git tag -a "${VERSION}" --file -
git push -q --follow-tags -u origin master > /dev/null 2>&1

cd ..
rm -Rf baqup-library-image/

echo "Deployed"
