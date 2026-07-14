#!/usr/bin/env bash
set -eo pipefail

ENV_FILE="${ENV_FILE:-$(pwd)/script/.env}"
ASSIMILIS_VERSION="${ASSIMILIS_VERSION:-v1.0.2}"
AIKIDO_BASE_URL="${AIKIDO_BASE_URL:-https://app.aikido.dev}"
SBOM_DIR="${SBOM_DIR:-$(pwd)/third_party/sbom}"
REPO_NAME="traefik"
ASSIMILIS_ERROR_LOG="$(pwd)/third_party/assimilis.log"
TIMESTAMP_FILE="$(pwd)/third_party/.last_generated_at"

if [[ ! -f "$ENV_FILE" ]]; then
    printf "Credentials file not found: ${ENV_FILE}
Create it with:

  cp .env.example .env

Then add your Aikido credentials."
    exit 1
fi

set -a
# shellcheck source=/dev/null
source "$ENV_FILE"
set +a

TMP_DIR=$(mktemp -d)

log() {
    printf '==> %s\n' "$*"
}

format_timestamp_utc() {
    local timestamp="$1"

    if date --version >/dev/null 2>&1; then
        # GNU date: Linux
        formatted_scan_time="$(
            date -u -d "@${timestamp}" '+%Y-%m-%d %H:%M:%S UTC'
        )"
    else
        # BSD date: macOS
        formatted_scan_time="$(
            date -u -r "${timestamp}" '+%Y-%m-%d %H:%M:%S UTC'
        )"
    fi
}

cleanup() {
  rm -rf "$TMP_DIR"
  rm -f "$ASSIMILIS_ERROR_LOG"
}

trap cleanup EXIT

log "Requesting access token from Aikido"

OAUTH_RESPONSE="${TMP_DIR}/oauth-response.json"

if ! oauth_http_code="$(
    curl -sS -u "${AIK_CLIENT}:${AIK_SECRET}" --request POST \
        --url ${AIKIDO_BASE_URL}/api/oauth/token \
        --write-out '%{http_code}' \
        --header 'accept: application/json' \
        --header 'content-type: application/json' \
        --data '{"grant_type": "client_credentials"}' \
        -o "${OAUTH_RESPONSE}"
)"; then
    printf "Could not connect to the Aikido authentication endpoint"
    exit 1
fi

if [[ ! "$oauth_http_code" =~ ^2[0-9][0-9]$ ]]; then
    printf "Aikido API error: $OAUTH_RESPONSE"
    printf "Aikido authentication failed with HTTP status ${oauth_http_code}"
    exit 1
fi

if ! AIKIDO_TOKEN="$(jq -er '.access_token' "${OAUTH_RESPONSE}")"; then
    printf "Failed to get access token from Aikido"
    exit 1
fi

log "Downloading the CycloneDX SBOM for Aikido repository ${AIKIDO_REPO_CODE}"

mkdir -p ${SBOM_DIR}
if ! sbom_http_code="$(
    curl -sS --request GET \
        --write-out '%{http_code}' \
        --url "${AIKIDO_BASE_URL}/api/public/v1/repositories/code/${AIKIDO_REPO_CODE}/licenses/export?format=sbom" \
        --header 'accept: application/json' \
        --header "authorization: Bearer ${AIKIDO_TOKEN}" \
        -o ${SBOM_DIR}/${REPO_NAME}.cdx.json
)"; then
    printf "Could not download the SBOM from Aikido"
    exit 1
fi

if [[ ! "$sbom_http_code" =~ ^2[0-9][0-9]$ ]]; then
    printf "Aikido API error: $SBOM_DIR/${REPO_NAME}.cdx.json"
    printf "Aikido SBOM export failed with HTTP status ${sbom_http_code}"
    exit 1
fi

log "SBOM saved to ${SBOM_DIR}/${REPO_NAME}.cdx.json"

REPOSITORY_RESPONSE="${TMP_DIR}/repository-response.json"

if repository_http_code="$(
    curl -sS --request GET \
        --url ${AIKIDO_BASE_URL}/api/public/v1/repositories/code/${AIKIDO_REPO_CODE} \
        --write-out '%{http_code}' \
        --header 'accept: application/json' \
        --header "authorization: Bearer ${AIKIDO_TOKEN}" \
        -o "${REPOSITORY_RESPONSE}"
)"; then
  if [[ "$repository_http_code" =~ ^2[0-9][0-9]$ ]]; then
    last_scanned_at="$(
      jq -er '.last_scanned_at' "$REPOSITORY_RESPONSE"
    )"
    if [[ -n "$last_scanned_at" ]]; then
      format_timestamp_utc "$last_scanned_at"
      log "Aikido repository last scanned at: ${formatted_scan_time}"
    else
      printf "Warning: Aikido did not provide a last_scanned_at value"
    fi
  else
    printf "Warning: Could not retrieve the Aikido scan time; HTTP status ${repository_http_code}"
  fi
else
  printf "Warning: Could not retrieve the Aikido repository metadata"
fi

log "Generating third-party attribution files"

set +e
assimilis --repo-name ${REPO_NAME} 2>&1 | tee "${ASSIMILIS_ERROR_LOG}"
assimilis_exit_code="${PIPESTATUS[0]}"
set -e

if [[ "$assimilis_exit_code" -ne 0 ]]; then
    printf "Error: Assimilis failed with exit code ${assimilis_exit_code}.

The complete log was saved to:
  ${ASSIMILIS_ERROR_LOG}"
    exit 1
fi

# Record when the attribution generation was completed successfully
generated_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
printf '%s\n' "$generated_at" > "$TIMESTAMP_FILE"

log "Third-party files generated successfully"
log "Generation timestamp: ${generated_at}"
