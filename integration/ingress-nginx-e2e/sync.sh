#!/usr/bin/env bash
# Vendors the ingress-nginx e2e annotation tests and framework, and re-applies the Traefik harness patch.
#
# Usage:
#   ./sync.sh          re-vendors the commit pinned in UPSTREAM and applies framework.patch.
#   ./sync.sh <sha>    same, but pins <sha> first.
#   ./sync.sh diff     regenerates framework.patch from the local modifications of vendored files.

set -eu -o pipefail

script_dir="$( cd "$( dirname "${0}" )" && pwd -P)"
cd "${script_dir}"

upstream_repo="https://github.com/kubernetes/ingress-nginx.git"

# Upstream paths copied as-is. The test packages must never be modified, only the framework is patched.
vendored_dirs=(
    "test/e2e/annotations"
    "test/e2e/framework"
)
vendored_files=(
    "test/e2e/e2e.go"
    "test/e2e/e2e_test.go"
    "test/e2e/HTTPBUN_IMAGE"
    "NGINX_BASE"
    "internal/k8s/main.go"
    "internal/k8s/zz_generated.deepcopy.go"
    "LICENSE"
)

checkout_upstream() {
    local sha="${1}" dest="${2}"
    mkdir -p "${dest}"
    git -C "${dest}" init --quiet
    git -C "${dest}" remote add origin "${upstream_repo}"
    git -C "${dest}" fetch --quiet --depth 1 origin "${sha}"
    git -C "${dest}" checkout --quiet FETCH_HEAD
}

copy_upstream() {
    local src="${1}" dest="${2}"
    for dir in "${vendored_dirs[@]}"; do
        mkdir -p "${dest}/${dir}"
        # The modsecurity annotation tests are a separate package that Traefik does not support.
        rsync -a --delete --exclude "modsecurity/" "${src}/${dir}/" "${dest}/${dir}/"
    done
    for file in "${vendored_files[@]}"; do
        mkdir -p "$(dirname "${dest}/${file}")"
        cp "${src}/${file}" "${dest}/${file}"
    done
}

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

if [ "${1:-}" = "diff" ]; then
    checkout_upstream "$(cat UPSTREAM)" "${tmp_dir}/upstream"
    mkdir -p "${tmp_dir}/a" "${tmp_dir}/b"
    copy_upstream "${tmp_dir}/upstream" "${tmp_dir}/a"
    copy_upstream "${script_dir}" "${tmp_dir}/b"
    # diff exits with 1 when files differ.
    (cd "${tmp_dir}" && diff -ruN a b > "${script_dir}/framework.patch") || true
    echo "framework.patch regenerated."
    exit 0
fi

if [ -n "${1:-}" ]; then
    echo "${1}" > UPSTREAM
fi

checkout_upstream "$(cat UPSTREAM)" "${tmp_dir}/upstream"
copy_upstream "${tmp_dir}/upstream" "${script_dir}"
patch --quiet -p1 < framework.patch
go mod tidy
echo "Vendored ingress-nginx $(cat UPSTREAM)."
