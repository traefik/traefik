#!/usr/bin/env bash
set -e

ci_retry make validate

if [ -n "$SHOULD_TEST" ]; then ci_retry make test-unit; fi

if [ -n "$SHOULD_TEST" ]; then make -j${N_MAKE_JOBS} crossbinary-default-parallel; fi
