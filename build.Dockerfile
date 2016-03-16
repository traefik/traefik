FROM golang:1.6.0-alpine

RUN apk update && apk add git bash gcc

RUN go get github.com/Masterminds/glide
RUN go get github.com/mitchellh/gox
RUN go get github.com/jteeuwen/go-bindata/...
RUN go get github.com/golang/lint/golint

# Which docker version to test on
ENV DOCKER_VERSION 1.10.1

# enable GO15VENDOREXPERIMENT
ENV GO15VENDOREXPERIMENT 1

# Download docker
RUN set -ex; \
    curl https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION} -o /usr/local/bin/docker-${DOCKER_VERSION}; \
    chmod +x /usr/local/bin/docker-${DOCKER_VERSION}

# Set the default Docker to be run
RUN ln -s /usr/local/bin/docker-${DOCKER_VERSION} /usr/local/bin/docker

WORKDIR /go/src/github.com/containous/traefik

COPY glide.yaml glide.yaml
COPY glide.lock glide.lock
RUN glide install

COPY . /go/src/github.com/containous/traefik
