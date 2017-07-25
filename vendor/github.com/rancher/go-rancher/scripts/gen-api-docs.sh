#!/bin/bash
set -e

cd $(dirname $0)/../docs

curl -s http://localhost/v1/schemas?_role=project | jq . > ./input/schemas.json
echo Saved schemas.json


 go run *.go -command=generate-collection-description

 go run *.go -command=generate-description

 go run *.go -command=generate-empty-description

 go run *.go -command=generate-only-resource-fields

 go run *.go -command=generate-docs -version=v1.2 -lang=en -layout=rancher-api-default-v1.2

echo Success
