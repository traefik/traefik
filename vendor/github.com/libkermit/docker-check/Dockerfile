FROM golang:1.8.3

RUN apt-get update && apt-get install -y \
    iptables build-essential \
    --no-install-recommends

# Install build dependencies
RUN go get golang.org/x/tools/cmd/cover \
    && go get github.com/golang/lint/golint \
    && go get github.com/rancher/trash 

# Which docker version to test on and what default one to use
ENV DOCKER_VERSION 17.03.2

# Download docker
RUN mkdir -p /usr/local/bin \
    && curl -fL https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}-ce.tgz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'
    
WORKDIR /go/src/github.com/libkermit/docker-check

COPY trash.yml .
RUN trash -k

# Wrap all commands in the "docker-in-docker" script to allow nested containers
ENTRYPOINT ["hack/dind"]

COPY . /go/src/github.com/libkermit/docker-check
RUN trash
