#!/usr/bin/env bash
set -e

if [ -n "${SEMAPHORE_GIT_TAG_NAME}" ]; then
  echo "Releasing packages..."
else
  echo "Skipping release"
  exit 0
fi

rm -rf dist

for os in linux darwin windows freebsd openbsd; do
    goreleaser release --skip=publish -p 2 --timeout="90m" --config "$(go run ./internal/release "$os")"
    go clean -cache
done

cat dist/**/*_checksums.txt >> "dist/traefik_${VERSION}_checksums.txt"
rm dist/**/*_checksums.txt
tar cfz "dist/traefik-${VERSION}.src.tar.gz" \
  --exclude-vcs \
  --exclude .idea \
  --exclude .travis \
  --exclude .semaphoreci \
  --exclude .github \
  --exclude dist .

chown -R "$(id -u)":"$(id -g)" dist/
