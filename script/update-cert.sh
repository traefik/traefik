#!/bin/sh

set -e

CERT_IMAGE="alpine:edge"

# cd to the current directory so the script can be run from anywhere.
cd `dirname $0`

# Update the cert image.
docker pull $CERT_IMAGE

# Fetch the latest certificates.
ID=$(docker run -d $CERT_IMAGE sh -c "apk --update upgrade && apk add ca-certificates && update-ca-certificates")
docker logs -f $ID
docker wait $ID

# Update the local certificates.
docker cp $ID:/etc/ssl/certs/ca-certificates.crt .

# Cleanup.
docker rm -f $ID
