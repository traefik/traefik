# This file describes the standard way to build libcompose, using docker
FROM golang:1.8.3

# virtualenv is necessary to run acceptance tests
RUN apt-get update && \
    apt-get install -y iptables build-essential --no-install-recommends && \
    apt-get install -y python-setuptools && \
    easy_install pip && pip install virtualenv

# Install build dependencies
RUN go get github.com/aktau/github-release && \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/golang/lint/golint

# Which docker version to test on and what default one to use
ENV DOCKER_VERSIONS 1.9.1 1.10.3 1.13.1 17.03.2 17.06.0
ENV DEFAULT_DOCKER_VERSION 17.03.2

# Download docker
RUN set -e; set -x; \
    for v in $(echo ${DOCKER_VERSIONS} | cut -f1); do \
        if test "${v}" = "1.9.1" || test "${v}" = "1.10.3"; then \
           mkdir -p /usr/local/bin/docker-${v}/; \
           curl https://get.docker.com/builds/Linux/x86_64/docker-${v} -o /usr/local/bin/docker-${v}/docker; \
           chmod +x /usr/local/bin/docker-${v}/docker; \
        elif test "${v}" = "1.13.1"; then \
           curl https://get.docker.com/builds/Linux/x86_64/docker-${v}.tgz -o docker-${v}.tgz; \
             tar xzf docker-${v}.tgz -C /usr/local/bin/; \
             mv /usr/local/bin/docker /usr/local/bin/docker-${v}; \
             rm docker-${v}.tgz; \
        else \
             curl https://download.docker.com/linux/static/stable/x86_64/docker-${v}-ce.tgz -o docker-${v}.tgz; \
             tar xzf docker-${v}.tgz -C /usr/local/bin/; \
             mv /usr/local/bin/docker /usr/local/bin/docker-${v}; \
             rm docker-${v}.tgz; \
        fi \
    done

# Set the default Docker to be run
RUN ln -s /usr/local/bin/docker-${DEFAULT_DOCKER_VERSION} /usr/local/bin/docker

WORKDIR /go/src/github.com/docker/libcompose

# Compose COMMIT for acceptance test version, update that commit when
# you want to update the acceptance test version to support.
ENV COMPOSE_COMMIT 1.9.0
RUN virtualenv venv && \
    git clone https://github.com/docker/compose.git venv/compose && \
    cd venv/compose && \
    git checkout -q "$COMPOSE_COMMIT" && \
    ../bin/pip install \
               -r requirements.txt \
               -r requirements-dev.txt

ENV COMPOSE_BINARY /go/src/github.com/docker/libcompose/libcompose-cli
ENV USER root

# Wrap all commands in the "docker-in-docker" script to allow nested containers
ENTRYPOINT ["hack/dind"]

COPY . /go/src/github.com/docker/libcompose
