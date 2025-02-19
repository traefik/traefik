#!/usr/bin/env bash

set -e -o pipefail

export PROJECT_MODULE="github.com/traefik/traefik"
export MODULE_VERSION="v2"
KUBE_VERSION=v0.30.10
CURRENT_DIR="$(pwd)"

# shellcheck disable=SC1091 # Cannot check source of this file
go install "k8s.io/code-generator/cmd/deepcopy-gen@${KUBE_VERSION}"

CODEGEN_PKG="$(go env GOPATH)/pkg/mod/k8s.io/code-generator@${KUBE_VERSION}"
source "${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_helpers \
  --boilerplate "$(dirname "${BASH_SOURCE[0]}")/boilerplate.go.tmpl" \
  "${CURRENT_DIR}"

kube::codegen::gen_client \
    --with-watch \
    --output-dir "${CURRENT_DIR}/pkg/provider/kubernetes/crd/generated" \
    --output-pkg "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd/generated" \
    --boilerplate "$(dirname "${BASH_SOURCE[0]}")/boilerplate.go.tmpl" \
    "${CURRENT_DIR}/pkg/provider/kubernetes/crd"
