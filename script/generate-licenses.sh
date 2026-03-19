#!/usr/bin/env bash
set -euo pipefail

# Generates SBOM and third-party attribution files using CycloneDX tools and assimilis.
#
# Tools used:
#   - cyclonedx-gomod: Go SBOM from go.mod (with test and stdlib dependencies)
#   - @cyclonedx/yarn-plugin-cyclonedx: npm SBOM via Yarn plugin (reads licenses from node_modules)
#   - cyclonedx-py: Python SBOM from docs/requirements.txt (via resolved venv)
#   - jq: Merges individual SBOMs into one
#   - assimilis: Generates human-readable attribution files from the merged SBOM

ASSIMILIS_VERSION="${ASSIMILIS_VERSION:-v1.0.2}"
CYCLONEDX_GOMOD_VERSION="${CYCLONEDX_GOMOD_VERSION:-v1.10.0}"
CYCLONEDX_PY_VERSION="${CYCLONEDX_PY_VERSION:-v8.7.0}"
REPO_NAME="${REPO_NAME:-traefik}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_DIR}"

# ─── Install tools if not already available ───

if ! command -v cyclonedx-gomod &>/dev/null; then
    echo "Installing cyclonedx-gomod ${CYCLONEDX_GOMOD_VERSION}..."
    go install "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@${CYCLONEDX_GOMOD_VERSION}"
fi

if ! command -v assimilis &>/dev/null; then
    echo "Installing assimilis ${ASSIMILIS_VERSION}..."
    GOBIN=/tmp/assimilis go install "github.com/traefik/assimilis/cmd@${ASSIMILIS_VERSION}"
    mv /tmp/assimilis/cmd "$(go env GOPATH)/bin/assimilis"
fi

if ! command -v jq &>/dev/null; then
    echo "Error: jq is required but not installed."
    exit 1
fi

if ! command -v cyclonedx-py &>/dev/null; then
    echo "Installing cyclonedx-py ${CYCLONEDX_PY_VERSION}..."
    if command -v pipx &>/dev/null; then
        pipx install "cyclonedx-bom==${CYCLONEDX_PY_VERSION#v}"
    elif command -v brew &>/dev/null; then
        brew install cyclonedx-py
    else
        echo "Error: cyclonedx-py is required. Install via pipx or brew."
        exit 1
    fi
fi

if ! command -v uv &>/dev/null; then
    echo "Installing uv..."
    if command -v brew &>/dev/null; then
        brew install uv
    else
        echo "Error: uv is required. Install via brew or https://docs.astral.sh/uv/"
        exit 1
    fi
fi

# ─── Prepare output directory ───

OUT_DIR="licenses"
WORK_DIR=$(mktemp -d)
trap 'rm -rf "${WORK_DIR}"' EXIT

mkdir -p "${WORK_DIR}/sbom"

# ─── Generate SBOMs in parallel ───

# Go SBOM.
echo "Generating Go SBOM (cyclonedx-gomod)..."
cyclonedx-gomod mod -test -std -json -licenses -assert-licenses \
    -output "${WORK_DIR}/sbom/go.cdx.json" &
go_pid=$!

# npm SBOM via Yarn CycloneDX plugin (reads licenses from node_modules/package.json).
echo "Generating npm SBOM (yarn cyclonedx)..."
(cd "${PROJECT_DIR}/webui" && yarn cyclonedx \
    --output-file "${WORK_DIR}/sbom/npm.cdx.json" 2>&1) &
npm_pid=$!

# Python SBOM from docs/requirements.txt.
echo "Generating Python SBOM (cyclonedx-py)..."
(
    venv_dir="${WORK_DIR}/venv"
    python3 -m venv "$venv_dir"
    uv pip compile "${PROJECT_DIR}/docs/requirements.txt" \
        -o "${venv_dir}/requirements-lock.txt" 2>/dev/null
    "$venv_dir/bin/pip" install -q -r "${venv_dir}/requirements-lock.txt"
    cyclonedx-py environment --of json \
        -o "${WORK_DIR}/sbom/py.cdx.json" \
        "$venv_dir" 2>&1
) &
py_pid=$!

# Wait for all generators.
wait "$go_pid" || { echo "Error: cyclonedx-gomod failed"; exit 1; }
wait "$npm_pid" || { echo "Error: yarn cyclonedx failed"; exit 1; }
wait "$py_pid" || { echo "Error: cyclonedx-py failed"; exit 1; }

# ─── Merge SBOMs ───

echo "Merging SBOMs..."
sboms_to_merge=()
[ -f "${WORK_DIR}/sbom/go.cdx.json" ] && sboms_to_merge+=("${WORK_DIR}/sbom/go.cdx.json")
[ -f "${WORK_DIR}/sbom/npm.cdx.json" ] && sboms_to_merge+=("${WORK_DIR}/sbom/npm.cdx.json")
[ -f "${WORK_DIR}/sbom/py.cdx.json" ] && sboms_to_merge+=("${WORK_DIR}/sbom/py.cdx.json")

if [ ${#sboms_to_merge[@]} -eq 0 ]; then
    echo "Error: no SBOMs were generated"
    exit 1
fi

jq -s '
    .[0] as $base |
    [.[] | .components // []] | add |
    unique_by(.purl // .name + "@" + (.version // "")) |
    $base + {components: .}
' "${sboms_to_merge[@]}" > "${WORK_DIR}/sbom/${REPO_NAME}.cdx.json"

echo "Merged SBOM: $(jq '.components | length' "${WORK_DIR}/sbom/${REPO_NAME}.cdx.json") components"

# ─── Copy merged SBOM to output and run assimilis ───

rm -rf "${OUT_DIR}/sbom"
mkdir -p "${OUT_DIR}/sbom" "${OUT_DIR}/licenses/custom"
cp "${WORK_DIR}/sbom/${REPO_NAME}.cdx.json" "${OUT_DIR}/sbom/${REPO_NAME}.cdx.json"

echo "Generating attributions..."
assimilis --repo-name "${REPO_NAME}" --output-dir "${OUT_DIR}"

echo "Done. Attribution files are in ${OUT_DIR}/."
