FROM golang:1.7

RUN go get github.com/jteeuwen/go-bindata/... \
&& go get github.com/golang/lint/golint \
&& go get github.com/kisielk/errcheck \
&& go get github.com/client9/misspell/cmd/misspell

# Which docker version to test on
ARG DOCKER_VERSION=1.10.3


# Which glide version to test on
ARG GLIDE_VERSION=v0.12.3

# Download glide
RUN mkdir -p /usr/local/bin \
    && curl -fL https://github.com/Masterminds/glide/releases/download/${GLIDE_VERSION}/glide-${GLIDE_VERSION}-linux-amd64.tar.gz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'

# Download docker
RUN mkdir -p /usr/local/bin \
    && curl -fL https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION}.tgz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'

WORKDIR /go/src/github.com/containous/traefik

COPY glide.yaml glide.yaml
COPY glide.lock glide.lock
RUN glide install -v

COPY integration/glide.yaml integration/glide.yaml
COPY integration/glide.lock integration/glide.lock
RUN cd integration && glide install

COPY . /go/src/github.com/containous/traefik
