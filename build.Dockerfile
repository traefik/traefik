FROM golang:1.5

RUN go get github.com/tools/godep
RUN go get github.com/mitchellh/gox
RUN go get github.com/tcnksm/ghr

# Which docker version to test on
ENV DOCKER_VERSION 1.6.2

# Download docker
RUN set -ex; \
    curl https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION} -o /usr/local/bin/docker-${DOCKER_VERSION}; \
    chmod +x /usr/local/bin/docker-${DOCKER_VERSION}

# Set the default Docker to be run
RUN ln -s /usr/local/bin/docker-${DOCKER_VERSION} /usr/local/bin/docker

ENV PATH /go/src/github.com/emilevauge/traefik/Godeps/_workspace/bin:$PATH

WORKDIR /go/src/github.com/emilevauge/traefik

# This is a hack (see libcompose#32) - will be removed when libcompose will be fixed
# (i.e go get able)
RUN mkdir -p /go/src/github.com/docker/docker/autogen/dockerversion/
COPY Godeps/_workspace/src/github.com/docker/docker/autogen/dockerversion/dockerversion.go /go/src/github.com/docker/docker/autogen/dockerversion/dockerversion.go

RUN mkdir Godeps
COPY Godeps/Godeps.json Godeps/
RUN godep restore

COPY . /go/src/github.com/emilevauge/traefik
