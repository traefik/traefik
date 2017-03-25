FROM golang:1.8

# Install a more recent version of mercurial to avoid mismatching results
# between glide run on a decently updated host system and the build container.
RUN awk '$1 ~ "^deb" { $3 = $3 "-backports"; print; exit }' /etc/apt/sources.list > /etc/apt/sources.list.d/backports.list && \
  DEBIAN_FRONTEND=noninteractive apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -t jessie-backports --yes --no-install-recommends mercurial=3.9.1-1~bpo8+1 && \
  rm -fr /var/lib/apt/lists/

RUN go get github.com/jteeuwen/go-bindata/... \
&& go get github.com/golang/lint/golint \
&& go get github.com/kisielk/errcheck \
&& go get github.com/client9/misspell/cmd/misspell \
&& go get github.com/mattfarina/glide-hash \
&& go get github.com/sgotti/glide-vc

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
COPY . /go/src/github.com/containous/traefik
