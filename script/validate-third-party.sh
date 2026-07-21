#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

TIMESTAMP_FILE="${TIMESTAMP_FILE:-${ROOT_DIR}/third_party/.last_generated_at}"
MAX_AGE="${MAX_AGE:-12h}"

die() { printf 'error: %s\n' "$*" >&2; exit 1; }

value="${MAX_AGE%?}"
unit="${MAX_AGE: -1}"

[[ "$value" =~ ^[1-9][0-9]*$ ]] ||
    die "MAX_AGE must be a positive duration such as 4h or 2d"

case "$unit" in
    m) max_age_seconds=$((value * 60)) ;;
    h) max_age_seconds=$((value * 60 * 60)) ;;
    d) max_age_seconds=$((value * 24 * 60 * 60)) ;;
    *) die "Unsupported MAX_AGE unit '${unit}'; use m, h, or d" ;;
esac

[[ -f "$TIMESTAMP_FILE" ]] ||
    die "Timestamp file not found: ${TIMESTAMP_FILE}. Run 'make third-party' first."

generated_at="$(tr -d '[:space:]' < "$TIMESTAMP_FILE")"

[[ "$generated_at" =~ ^[0-9]+$ ]] ||
    die "Invalid timestamp in ${TIMESTAMP_FILE}"

current_time="$(date -u '+%s')"
age=$((current_time - generated_at))

(( age >= 0 )) || die "Third-party generation timestamp is in the future"
(( age <= max_age_seconds )) || die "Third-party files are too old. Run 'make third-party' before releasing."

printf 'Third-party files are recent enough for this release.\n'