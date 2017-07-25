#!/bin/bash

# This script is designed to be run from ../Makefile via `make docker-minimal-linux`.
# $1 - Seal key to build with
#
# Overridable variables:
# GO_PROJECT_PATH: go import path for the project
# DOCKER_IMAGE_NAME: name for final docker image
# DOCKER_IMAGE_TAG: tag for final docker image

GO_PROJECT_PATH=${GO_PROJECT_PATH-github.com/vulcand/vulcand}

echo "Building go project \"$GO_PROJECT_PATH\""

cp -a ${HOST_PROJECT_PATH} ${GOPATH}/src/
go get vulcand
go get vulcand/vctl
go get vulcand/vbundle

# Fixes for statically compling on go1.4 - https://github.com/golang/go/issues/9344
CGO_ENABLED=0 go build -a -tags netgo -installsuffix cgo --ldflags '-extldflags \"-static\" -s -w' ${GO_PROJECT_PATH}
CGO_ENABLED=0 go build -a -tags netgo -installsuffix cgo --ldflags '-extldflags \"-static\" -s -w' ${GO_PROJECT_PATH}/vctl
CGO_ENABLED=0 go build -a -tags netgo -installsuffix cgo --ldflags '-extldflags \"-static\" -s -w' ${GO_PROJECT_PATH}/vbundle

cat > Dockerfile-minimal << EOF
FROM scratch
EXPOSE 8181 8182
COPY vulcand /app/vulcand
COPY vctl /app/vctl
COPY vbundle /app/vbundle
ENV PATH=/app:$PATH

ENTRYPOINT ["/app/vulcand"]
CMD ["-etcd=http://127.0.0.1:4001", "-etcd=127.0.0.1:4002", "-etcd=127.0.0.1:4003", "-sealKey=$1"]
EOF

docker build --no-cache -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG-latest} -f Dockerfile-minimal .
