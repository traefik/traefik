#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

####
### Helper script for glide[-vc] to handle specifics for the Traefik repo.
##
# In particular, the 'integration/' directory contains its own set of
# glide-managed dependencies which must not have its nested vendor folder
# stripped. Depending on where the script is called from, it will do the Right
# Thing.
#

CWD="$(pwd)"; readonly CWD
GLIDE_ARGS=()
GLIDE_VC_ARGS=(
  '--use-lock-file'       # `glide list` seems to miss test dependencies, e.g., github.com/mattn/go-shellwords
  '--only-code'
  '--no-tests'
)

usage() {
  echo "usage: $(basename "$0") install | update | get <package> | trim
install: Install all dependencies and trim the vendor folder afterwards (alternative command: i).
update:  Update all dependencies and trim the vendor folder afterwards (alternative command: up).
get:     Add a dependency and trim the vendor folder afterwards.
trim:    Trim the vendor folder only, do not install or update dependencies.

The current working directory must contain a glide.yaml file." >&2
}

is_integration_dir() {
  [[ "$(basename ${CWD})" = 'integration' ]]
}

if ! is_integration_dir; then
  GLIDE_ARGS+=('--strip-vendor' '--skip-test')
fi

if ! type glide > /dev/null 2>&1; then
  echo "glide not found in PATH." >&2
  exit 1
fi

if ! type glide-vc > /dev/null 2>&1; then
  echo "glide-vc not found in PATH." >&2
  exit 1
fi

if [[ ! -e "${CWD}/glide.yaml" ]]; then
  echo "no glide.yaml file found in the current working directory" >&2
  exit 1
fi

if [[ $# -lt 1 ]]; then
  echo "missing command" >&2
  usage
  exit 1
fi

readonly glide_command="$1"
shift

skip_glide_command=
case "${glide_command}" in
  'install' | 'i')
    if [[ $# -ne 0 ]]; then
      echo "surplus parameters given" >&2
      usage
      exit 1
    fi
    ;;

  'update' | 'up')
    if [[ $# -ne 0 ]]; then
      echo "surplus parameters given" >&2
      usage
      exit 1
    fi
    ;;

  'get')
    if [[ $# -ne 1 ]]; then
      echo 'insufficient/surplus arguments given for "get" command' >&2
      usage
      exit 1
    fi
    GLIDE_ARGS+=("$1")
    shift
    ;;

  'trim')
    if [[ $# -ne 0 ]]; then
      echo "surplus parameters given" >&2
      usage
      exit 1
    fi
    skip_glide_command=yes
    ;;

  *)
    echo "unknown command: ${glide_command}" >&2
    usage
    exit 1
esac
readonly skip_glide_command

if [[ -z "${skip_glide_command}" ]]; then
  # Use parameter substitution to account for an empty glide arguments array
  # that would otherwise lead to an "unbound variable" error due to the nounset
  # option.
  GLIDE_ARGS=("${GLIDE_ARGS+"${GLIDE_ARGS[@]}"}")
  echo "running: glide ${glide_command} ${GLIDE_ARGS[*]}"
  glide ${glide_command} ${GLIDE_ARGS[*]}
fi

echo "trimming vendor folder using: glide-vc ${GLIDE_VC_ARGS[*]}"
glide-vc ${GLIDE_VC_ARGS[*]}
