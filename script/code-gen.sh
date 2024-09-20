#!/usr/bin/env bash

set -e -o pipefail

# shellcheck disable=SC1091 # Cannot check source of this file
source /go/src/k8s.io/code-generator/kube_codegen.sh

git config --global --add safe.directory "/go/src/${PROJECT_MODULE}"

rm -rf "/go/src/${PROJECT_MODULE}/${MODULE_VERSION}"
mkdir -p "/go/src/${PROJECT_MODULE}/${MODULE_VERSION}/"

# TODO: remove the workaround when the issue is solved in the code-generator
# (https://github.com/kubernetes/code-generator/issues/165).
# Here, we create the soft link named "${PROJECT_MODULE}" to the parent directory of
# Traefik to ensure the layout required by the kube_codegen.sh script.
ln -s "/go/src/${PROJECT_MODULE}/pkg" "/go/src/${PROJECT_MODULE}/${MODULE_VERSION}/"

kube::codegen::gen_helpers \
    --input-pkg-root "${PROJECT_MODULE}/pkg" \
    --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
    --boilerplate "/go/src/${PROJECT_MODULE}/script/boilerplate.go.tmpl"

kube::codegen::gen_client \
    --with-watch \
    --input-pkg-root "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd" \
    --output-pkg-root "${PROJECT_MODULE}/${MODULE_VERSION}/pkg/provider/kubernetes/crd/generated" \
    --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
    --boilerplate "/go/src/${PROJECT_MODULE}/script/boilerplate.go.tmpl"

rm -rf "/go/src/${PROJECT_MODULE}/${MODULE_VERSION}"
