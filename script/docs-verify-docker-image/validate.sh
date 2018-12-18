#!/bin/sh

set -eu

PATH_TO_SITE="/app/site"
[ -d "${PATH_TO_SITE}" ]

NUMBER_OF_CPUS="$(grep -c processor /proc/cpuinfo)"


echo "=== Checking HTML content..."

# Search for all HTML files except the theme's partials
# and pipe this to htmlproofer with parallel threads
# (one htmlproofer per vCPU)
find "${PATH_TO_SITE}" -type f -not -path "/app/site/theme/*" \
    -name "*.html" -print0 \
| xargs -0 -r -P "${NUMBER_OF_CPUS}" -I '{}' \
  htmlproofer \
  --check-html \
  --check_external_hash \
  --alt_ignore="/traefik.logo.png/" \
  --http_status_ignore="0,500,501,503" \
  --url-ignore "/https://groups.google.com/a/traefik.io/forum/#!forum/security/,/localhost:/,/127.0.0.1:/,/fonts.gstatic.com/,/.minikube/,/github.com\/containous\/traefik\/*edit*/,/github.com\/containous\/traefik\/$/" \
  '{}'
## HTML-proofer options at https://github.com/gjtorikian/html-proofer#configuration

echo "= Documentation checked successfuly."
