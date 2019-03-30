FROM golang:1.11-alpine

RUN apk --update upgrade \
    && apk --no-cache --no-progress add git mercurial bash gcc musl-dev curl tar ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

RUN go get golang.org/x/lint/golint \
&& go get github.com/kisielk/errcheck \
&& go get github.com/client9/misspell/cmd/misspell

# Which docker version to test on
ARG DOCKER_VERSION=17.03.2
ARG DEP_VERSION=0.5.1

# Download go-bindata binary to bin folder in $GOPATH
RUN mkdir -p /usr/local/bin \
    && curl -fsSL -o /usr/local/bin/go-bindata https://github.com/containous/go-bindata/releases/download/v1.0.0/go-bindata \
    && chmod +x /usr/local/bin/go-bindata

# Download dep binary to bin folder in $GOPATH
RUN mkdir -p /usr/local/bin \
    && curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep

# Download docker
RUN mkdir -p /usr/local/bin \
    && curl -fL https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}-ce.tgz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'

WORKDIR /go/src/github.com/containous/traefik
COPY . /go/src/github.com/containous/traefik
