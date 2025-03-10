#!/usr/bin/env bash

set -e -o pipefail

PROJECT_MODULE="github.com/traefik/traefik"
MODULE_VERSION="v3"
KUBE_VERSION=v0.30.10
CURRENT_DIR="$(pwd)"

go install "k8s.io/code-generator/cmd/deepcopy-gen@${KUBE_VERSION}"
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.1

CODEGEN_PKG="$(go env GOPATH)/pkg/mod/k8s.io/code-generator@${KUBE_VERSION}"
# shellcheck disable=SC1091 # Cannot check source of this file
source "${CODEGEN_PKG}/kube_codegen.sh"

echo "# Generating Traefik clientset and deepcopy code ..."
kube::codegen::gen_helpers \
  --boilerplate "$(dirname "${BASH_SOURCE[0]}")/boilerplate.go.tmpl" \
  "${CURRENT_DIR}"

kube::codegen::gen_client \
    --with-watch \
    --output-dir "${CURRENT_DIR}/pkg/provider/kubernetes/crd/generated" \
    --output-pkg "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd/generated" \
    --boilerplate "$(dirname "${BASH_SOURCE[0]}")/boilerplate.go.tmpl" \
    "${CURRENT_DIR}/pkg/provider/kubernetes/crd"

echo "# Generating the CRD definitions for the documentation ..."
controller-gen crd:crdVersions=v1 \
    paths={./pkg/provider/kubernetes/crd/traefikio/v1alpha1/...} \
    output:dir=./docs/content/reference/dynamic-configuration/

echo "# Concatenate the CRD definitions for publication and integration tests ..."
cat "${CURRENT_DIR}"/docs/content/reference/dynamic-configuration/traefik.io_*.yaml > "${CURRENT_DIR}"/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml
cp -f "${CURRENT_DIR}"/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml "${CURRENT_DIR}"/integration/fixtures/k8s/01-traefik-crd.yml
