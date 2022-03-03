FROM golang:1.17-alpine

RUN apk --no-cache --no-progress add git mercurial bash gcc musl-dev curl tar ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

# Which docker version to test on
ARG DOCKER_VERSION=18.09.7

# Download docker
RUN mkdir -p /usr/local/bin \
    && curl -fL https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'

# Download golangci-lint binary to bin folder in $GOPATH
RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b $GOPATH/bin v1.44.0

# Download misspell binary to bin folder in $GOPATH
RUN curl -sfL https://raw.githubusercontent.com/client9/misspell/master/install-misspell.sh | bash -s -- -b $GOPATH/bin v0.3.4

# Download goreleaser binary to bin folder in $GOPATH
RUN curl -sfL https://gist.githubusercontent.com/traefiker/6d7ac019c11d011e4f131bb2cca8900e/raw/goreleaser.sh | sh

WORKDIR /go/src/github.com/traefik/traefik

# Download go modules
COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on GOPROXY=https://proxy.golang.org go mod download

COPY . /go/src/github.com/traefik/traefik
