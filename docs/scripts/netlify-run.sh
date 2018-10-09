#!/bin/sh
#
# This script is run in netlify environment to build and validate
# the website for documentation

set -eu

CURRENT_DIR="$(cd "$(dirname "${0}")" && pwd -P)"

#### Prepare environment
pip install -r requirements.txt

#### Build website
# Provide the URL for this deployment to Mkdocs
echo "${DEPLOY_PRIME_URL}" > "${CURRENT_DIR}/../CNAME"
sed -i "s#site_url:.*#site_url: ${DEPLOY_PRIME_URL}#" "${CURRENT_DIR}/../mkdocs.yml"

# Build
mkdocs build

exit 0
