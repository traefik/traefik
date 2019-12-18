ARG TRAEFIK_IMAGE_VERSION=dev

# ------------------------------------------------------------------------------

FROM traefik-frontend:$TRAEFIK_IMAGE_VERSION AS frontend

# ------------------------------------------------------------------------------

FROM golang:1.13-alpine

RUN apk --update upgrade \
    && apk --no-cache --no-progress add git mercurial bash gcc musl-dev curl tar ca-certificates tzdata make \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

# Download go-bindata binary to bin folder in $GOPATH
RUN mkdir -p /usr/local/bin \
    && curl -fsSL -o /usr/local/bin/go-bindata https://github.com/containous/go-bindata/releases/download/v1.0.0/go-bindata \
    && chmod +x /usr/local/bin/go-bindata

# Download golangci-lint binary to bin folder in $GOPATH
RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $GOPATH/bin v1.20.0

# Download golangci-lint and misspell binary to bin folder in $GOPATH
RUN GO111MODULE=off go get github.com/client9/misspell/cmd/misspell

# Download goreleaser binary to bin folder in $GOPATH
RUN curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh

WORKDIR /go/src/github.com/containous/traefik

COPY . /go/src/github.com/containous/traefik

RUN rm -rf static/*
COPY --from=frontend /src/webui/dist/pwa/ /go/src/github.com/containous/traefik/static/

# Download go modules
RUN GO111MODULE=on GOPROXY=https://proxy.golang.org go mod download

RUN ./script/make.sh generate binary
