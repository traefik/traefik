FROM golang:1.5.3

RUN go get github.com/Masterminds/glide
RUN go get github.com/mitchellh/gox
RUN go get github.com/tcnksm/ghr
RUN go get github.com/jteeuwen/go-bindata/...
RUN go get github.com/golang/lint/golint

# Which docker version to test on
ENV DOCKER_VERSION 1.6.2

# enable GO15VENDOREXPERIMENT
ENV GO15VENDOREXPERIMENT 1

# Download docker
RUN set -ex; \
    curl https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION} -o /usr/local/bin/docker-${DOCKER_VERSION}; \
    chmod +x /usr/local/bin/docker-${DOCKER_VERSION}

# Set the default Docker to be run
RUN ln -s /usr/local/bin/docker-${DOCKER_VERSION} /usr/local/bin/docker

WORKDIR /go/src/github.com/emilevauge/traefik

COPY glide.yaml glide.yaml
RUN glide up --quick

COPY . /go/src/github.com/emilevauge/traefik
