#!/usr/bin/env bash
set -e

make test-unit
ci_retry make test-integration
make -j${N_MAKE_JOBS} crossbinary-default-parallel
