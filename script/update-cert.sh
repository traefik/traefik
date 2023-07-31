#!/bin/sh

set -e

CERT_IMAGE="alpine:edge"

# cd to the current directory so the script can be run from anywhere.
SCRIPT_DIR="$( cd "$( dirname "${0}" )" && pwd -P)"; export SCRIPT_DIR
cd "${SCRIPT_DIR}"

# Update the cert image.
/usr/bin/docker pull $CERT_IMAGE

# Fetch the latest certificates.
ID=$(/usr/bin/docker run -d $CERT_IMAGE sh -c "apk --update upgrade && apk add ca-certificates && update-ca-certificates")
/usr/bin/docker logs -f "${ID}"
/usr/bin/docker wait "${ID}"

# Update the local certificates.
/usr/bin/docker cp "${ID}":/etc/ssl/certs/ca-certificates.crt "${SCRIPT_DIR}"

# Cleanup.
/usr/bin/docker rm -f "${ID}"
