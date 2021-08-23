#!/bin/bash -e

set -e -o pipefail

HACK_DIR="$( cd "$( dirname "${0}" )" && pwd -P)"; export HACK_DIR
REPO_ROOT=${HACK_DIR}/..

# Generate the CRD definitions for the documentation
rm -rf "${REPO_ROOT}"/vendor
go mod vendor

go run "${REPO_ROOT}"/vendor/sigs.k8s.io/controller-tools/cmd/controller-gen \
  crd:crdVersions=v1 \
  paths="${REPO_ROOT}"/pkg/provider/kubernetes/crd/traefik/v1alpha1/... \
  output:dir="${REPO_ROOT}"/docs/content/reference/dynamic-configuration/

# Concatenate the CRD definitions for the integration tests
cat "${REPO_ROOT}"/docs/content/reference/dynamic-configuration/traefik.containo.us_*.yaml > "${REPO_ROOT}"/integration/fixtures/k8s/01-traefik-crd.yml

rm -rf "${REPO_ROOT}"/vendor
