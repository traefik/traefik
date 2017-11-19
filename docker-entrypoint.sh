#!/bin/sh
set -e

# make traefik command line options from environment variables starting with 'TRAEFIK_'
options=""
while IFS= read -r line
do
  var="${line%%=*}"
  if [ "${var#TRAEFIK_}" != "${var}" ]; then
    option=$(echo "--${var#TRAEFIK_}" | tr _ . | tr '[:upper:]' '[:lower:]')
    options="${options} ${option}"

    value="${line#${var}}"
    if [ "$value" != "=" ]; then
        options="${options}${value}"
    fi
  fi
done << EOF
$(env)
EOF

# this if will check if the first argument is a flag
# but only works if all arguments require a hyphenated flag
# -v; -SL; -f arg; etc will work, but not arg1 arg2
if [ "${1:0:1}" = '-' ]; then
    set -- /traefik $options "$@"
fi

# check for the expected command
if [ "$1" = 'traefik' ]; then
    exec /traefik $options "$@"
fi

# check for the debugoptions command
if [ "$1" = 'debugoptions' ]; then
    shift
    set -- /traefik $options "$@"
    echo "would execute: $@"
    exit
fi

# else default to run whatever the user wanted like "bash"
exec "$@"