#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

ENV_FILE="${ENV_FILE:-${SCRIPT_DIR}/.env}"
ASSIMILIS_VERSION="${ASSIMILIS_VERSION:-v1.0.2}"
AIKIDO_BASE_URL="${AIKIDO_BASE_URL:-https://app.aikido.dev}"
SBOM_DIR="${SBOM_DIR:-${ROOT_DIR}/third_party/sbom}"
TIMESTAMP_FILE="${ROOT_DIR}/third_party/.last_generated_at"
REPO_NAME="traefik"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

log() { printf '==> %s\n' "$*"; }

die() { printf 'error: %s\n' "$*" >&2; exit 1; }

load_env() {
    if [[ ! -f "$ENV_FILE" ]]; then
        die "Credentials file not found: ${ENV_FILE}

Create it with:

  cp script/.env.example script/.env

Then add your Aikido credentials."
    fi

    set -a
    # shellcheck source=/dev/null
    source "$ENV_FILE"
    set +a

    [[ -n "${AIK_CLIENT:-}" ]] || die "AIK_CLIENT is not set in ${ENV_FILE}"
    [[ -n "${AIK_SECRET:-}" ]] || die "AIK_SECRET is not set in ${ENV_FILE}"
    [[ -n "${AIKIDO_REPO_CODE:-}" ]] || die "AIKIDO_REPO_CODE is not set in ${ENV_FILE}"
}

aikido_api() {
    local method="$1"
    local path="$2"
    local body="${TMP_DIR}/response"
    local code

    shift 2

    code="$(
        curl -sS -X "$method" "${AIKIDO_BASE_URL}${path}" \
            -w '%{http_code}' -o "$body" "$@")" || die "Cannot reach Aikido (${path})"

    if [[ ! "$code" =~ ^2[0-9][0-9]$ ]]; then
        die "Aikido ${path} returned HTTP ${code}: $(cat "$body")"
    fi

    cat "$body"
}

get_token() {
    aikido_api POST /api/oauth/token \
        --user "${AIK_CLIENT}:${AIK_SECRET}" \
        --header 'accept: application/json' \
        --header 'content-type: application/json' \
        --data '{"grant_type":"client_credentials"}' |
        jq -er '.access_token | strings | select(length > 0)'
}

download_sbom() {
    local token="$1"
    local output="${SBOM_DIR}/${REPO_NAME}.cdx.json"
    local temporary_sbom="${TMP_DIR}/sbom.json"

    mkdir -p "$SBOM_DIR"

    aikido_api GET \
        "/api/public/v1/repositories/code/${AIKIDO_REPO_CODE}/licenses/export?format=sbom" \
        --header 'accept: application/json' \
        --header "authorization: Bearer ${token}" \
        > "$temporary_sbom"

    mv "$temporary_sbom" "$output"

    log "SBOM saved to ${output}"
}

# TODO: update with new Assimilis release
install_assimilis() {
    local bin_dir="${TMP_DIR}/assimilis-bin"

    mkdir -p "$bin_dir"

    log "Installing Assimilis ${ASSIMILIS_VERSION}"

    GOBIN="$bin_dir" go install \
        "github.com/traefik/assimilis/cmd@${ASSIMILIS_VERSION}"

    ASSIMILIS_BIN="${bin_dir}/cmd"

    [[ -x "$ASSIMILIS_BIN" ]] ||
        die "Assimilis executable was not created"
}

main() {
    local token
    local generated_at

    load_env

    log "Requesting access token from Aikido"

    token="$(get_token)" ||
        die "Aikido response did not contain an access token"

    log "Downloading the CycloneDX SBOM for Aikido repository ${AIKIDO_REPO_CODE}"

    download_sbom "$token"
    # install_assimilis

    log "Generating third-party attribution files"

    if ! (
        cd "$ROOT_DIR"
        # "$ASSIMILIS_BIN" --repo-name "$REPO_NAME"
        assimilis --repo-name "$REPO_NAME"
    ); then
        die "Assimilis failed"
    fi

    date -u '+%s' > "$TIMESTAMP_FILE"

    log "Third-party files generated successfully"
    log "Generation timestamp: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
}

main "$@"