#!/usr/bin/env bash
set -e

if [ -n "$TRAVIS" ]; then
  # Install all packages - this will make running the suite faster
  echo "--> Installing packages for faster tests"
  go install -tags="${GOTAGS}" -a ./...
fi

# If we are testing the API, build and install consul
if grep -q "/consul/api" <<< "${GOFILES}"; then
  # Create a temp dir and clean it up on exit
  TEMPDIR=`mktemp -d -t consul-test.XXX`
  trap "rm -rf ${TEMPDIR}" EXIT HUP INT QUIT TERM

  # Build the Consul binary for the API tests
  echo "--> Building consul"
  go build -tags="${GOTAGS}" -o $TEMPDIR/consul
  PATH="${TEMPDIR}:${PATH}"
fi

# Run the tests
echo "--> Running tests"
go test -timeout=360s -tags="${GOTAGS}" ${GOFILES} ${TESTARGS}
