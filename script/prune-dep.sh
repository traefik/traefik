#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

echo "Prune dependencies"

find vendor -name '*_test.go' -exec rm {} \;

find vendor -type f  \( ! -iname 'licen[cs]e*' \
 -a ! -iname '*notice*' \
 -a ! -iname '*patent*' \
 -a ! -iname '*copying*' \
 -a ! -iname '*unlicense*' \
 -a ! -iname '*copyright*' \
 -a ! -iname '*copyleft*' \
 -a ! -iname '*legal*' \
 -a ! -iname 'disclaimer*' \
 -a ! -iname 'third-party*' \
 -a ! -iname 'thirdparty*' \
 -a ! -iname '*.go' \
 -a ! -iname '*.c' \
 -a ! -iname '*.s' \
 -a ! -iname '*.pl' \
 -a ! -iname '*.cc' \
 -a ! -iname '*.cpp' \
 -a ! -iname '*.cxx' \
 -a ! -iname '*.h' \
 -a ! -iname '*.hh' \
 -a ! -iname '*.hpp' \
 -a ! -iname '*.hxx' \
 -a ! -iname '*.s' \) -exec rm -f {} +

find . -type d \( -iname '*Godeps*' \) -exec rm -rf {} +

find vendor -type l  \( ! -iname 'licen[cs]e*' \
 -a ! -iname '*notice*' \
 -a ! -iname '*patent*' \
 -a ! -iname '*copying*' \
 -a ! -iname '*unlicense*' \
 -a ! -iname '*copyright*' \
 -a ! -iname '*copyleft*' \
 -a ! -iname '*legal*' \
 -a ! -iname 'disclaimer*' \
 -a ! -iname 'third-party*' \
 -a ! -iname 'thirdparty*' \
 -a ! -iname '*.go' \
 -a ! -iname '*.c' \
 -a ! -iname '*.S' \
 -a ! -iname '*.cc' \
 -a ! -iname '*.cpp' \
 -a ! -iname '*.cxx' \
 -a ! -iname '*.h' \
 -a ! -iname '*.hh' \
 -a ! -iname '*.hpp' \
 -a ! -iname '*.hxx' \
 -a ! -iname '*.s' \) -exec rm -f {} +
