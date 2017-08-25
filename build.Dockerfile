FROM golang:1.9-alpine

RUN apk --update upgrade \
&& apk --no-cache --no-progress add git mercurial bash gcc musl-dev curl tar \
&& rm -rf /var/cache/apk/*

RUN go get github.com/jteeuwen/go-bindata/... \
&& go get github.com/golang/lint/golint \
&& go get github.com/kisielk/errcheck \
&& go get github.com/client9/misspell/cmd/misspell \
&& go get github.com/mattfarina/glide-hash \
&& go get github.com/sgotti/glide-vc

# Which docker version to test on
ARG DOCKER_VERSION=17.03.2

# Which glide version to test on
ARG GLIDE_VERSION=v0.12.3

# Download glide
RUN mkdir -p /usr/local/bin \
    && curl -fL https://github.com/Masterminds/glide/releases/download/${GLIDE_VERSION}/glide-${GLIDE_VERSION}-linux-amd64.tar.gz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'

# Download docker
RUN mkdir -p /usr/local/bin \
    && curl -fL https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}-ce.tgz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'

WORKDIR /go/src/github.com/containous/traefik
COPY . /go/src/github.com/containous/traefik
