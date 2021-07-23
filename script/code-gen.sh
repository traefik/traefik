#!/bin/bash -e

set -e -o pipefail

PROJECT_MODULE="github.com/traefik/traefik"
MODULE_VERSION="v2"
IMAGE_NAME="kubernetes-codegen:latest"

echo "Building codegen Docker image..."
docker build --build-arg KUBE_VERSION=v0.20.2 --build-arg USER=$USER --build-arg UID=$(id -u) --build-arg GID=$(id -g) -f "./script/codegen.Dockerfile" \
             -t "${IMAGE_NAME}" \
             "."

cmd="/go/src/k8s.io/code-generator/generate-groups.sh all ${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd/generated ${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd traefik:v1alpha1 --go-header-file=/go/src/${PROJECT_MODULE}/script/boilerplate.go.tmpl"

echo "Generating Traefik clientSet code ..."
echo $(pwd)
docker run --rm \
           -v "$(pwd):/go/src/${PROJECT_MODULE}" \
           -w "/go/src/${PROJECT_MODULE}" \
           "${IMAGE_NAME}" $cmd

cp -r $(pwd)/${MODULE_VERSION}/* $(pwd)
rm -rf $(pwd)/${MODULE_VERSION}
