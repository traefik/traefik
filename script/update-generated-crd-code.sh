#!/bin/bash -e

HACK_DIR=$(dirname "${BASH_SOURCE}")
REPO_ROOT=${HACK_DIR}/..

${REPO_ROOT}/vendor/k8s.io/code-generator/generate-groups.sh \
  all \
  github.com/containous/traefik/pkg/provider/kubernetes/crd/generated \
  github.com/containous/traefik/pkg/provider/kubernetes/crd \
  traefik:v1alpha1 \
  --go-header-file ${HACK_DIR}/boilerplate.go.tmpl \
  $@

deepcopy-gen  --input-dirs github.com/containous/traefik/pkg/config  -O zz_generated.deepcopy --go-header-file ${HACK_DIR}/boilerplate.go.tmpl
