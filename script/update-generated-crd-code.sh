#!/bin/bash -e

HACK_DIR="$( cd "$( dirname "${0}" )" && pwd -P)"; export HACK_DIR
REPO_ROOT=${HACK_DIR}/..

rm -rf "${REPO_ROOT}"/vendor
go mod vendor
chmod +x "${REPO_ROOT}"/vendor/k8s.io/code-generator/*.sh

"${REPO_ROOT}"/vendor/k8s.io/code-generator/generate-groups.sh \
  all \
  github.com/containous/traefik/pkg/provider/kubernetes/crd/generated \
  github.com/containous/traefik/pkg/provider/kubernetes/crd \
  traefik:v1alpha1 \
  --go-header-file "${HACK_DIR}"/boilerplate.go.tmpl \
  "$@"

deepcopy-gen \
--input-dirs github.com/containous/traefik/pkg/config/dynamic \
--input-dirs github.com/containous/traefik/pkg/tls \
--input-dirs github.com/containous/traefik/pkg/types \
-O zz_generated.deepcopy --go-header-file "${HACK_DIR}"/boilerplate.go.tmpl

rm -rf "${REPO_ROOT}"/vendor
