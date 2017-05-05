#!/usr/bin/env bash
set -e

make test-unit && make test-integration
make -j${N_MAKE_JOBS} crossbinary-default-parallel
