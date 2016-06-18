FROM golang:1.6.2

RUN go get github.com/Masterminds/glide \
&& go get github.com/jteeuwen/go-bindata/... \
&& go get github.com/golang/lint/golint \
&& go get github.com/kisielk/errcheck

# Which docker version to test on
ARG DOCKER_VERSION=1.10.1

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

RUN apt-get update && apt-get install -y bzip2 && apt-get clean && rm -rf /var/lib/apt/lists/*
RUN curl -sL http://upx.sourceforge.net/download/upx-3.91-amd64_linux.tar.bz2 | tar -jxv -C/usr/bin --strip-components=1

COPY . /go/src/github.com/containous/traefik
