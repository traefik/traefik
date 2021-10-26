#!/bin/bash -e

set -e -o pipefail

PROJECT_MODULE="github.com/traefik/traefik"
MODULE_VERSION="v2"
IMAGE_NAME="kubernetes-codegen:latest"

echo "Building codegen Docker image..."
docker build --build-arg KUBE_VERSION=v0.20.2 --build-arg USER="${USER}" --build-arg UID="$(id -u)" --build-arg GID="$(id -g)" -f "./script/codegen.Dockerfile" \
             -t "${IMAGE_NAME}" \
             "."

echo "Generating Traefik clientSet code ..."
docker run --rm \
           -v "${PWD}:/go/src/${PROJECT_MODULE}" \
           -w "/go/src/${PROJECT_MODULE}" \
           "${IMAGE_NAME}" \
           /go/src/k8s.io/code-generator/generate-groups.sh all "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd/generated" "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd" traefik:v1alpha1 --go-header-file="/go/src/${PROJECT_MODULE}/script/boilerplate.go.tmpl"

echo "Generating DeepCopy code ..."
docker run --rm \
           -v "${PWD}:/go/src/${PROJECT_MODULE}" \
           -w "/go/src/${PROJECT_MODULE}" \
           "${IMAGE_NAME}" \
           deepcopy-gen --input-dirs "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/config/dynamic" --input-dirs "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/tls" --input-dirs "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/types" --output-package "${PROJECT_MODULE}/${MODULE_VERSION}" -O zz_generated.deepcopy --go-header-file="/go/src/${PROJECT_MODULE}/script/boilerplate.go.tmpl"

echo "Generating the CRD definitions for the documentation ..."
docker run --rm \
           -v "${PWD}:/go/src/${PROJECT_MODULE}" \
           -w "/go/src/${PROJECT_MODULE}" \
           "${IMAGE_NAME}" \
           controller-gen crd:crdVersions=v1 paths=./pkg/provider/kubernetes/crd/traefik/v1alpha1/... output:dir=./docs/content/reference/dynamic-configuration/

echo "Concatenate the CRD definitions for the integration tests ..."
cat "${PWD}/docs/content/reference/dynamic-configuration/traefik.containo.us_*.yaml" > "${PWD}/integration/fixtures/k8s/01-traefik-crd.yml"

cp -r "${PWD}/${MODULE_VERSION}/*" "${PWD}"
rm -rf "${PWD}/${MODULE_VERSION:?}"
