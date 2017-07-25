FROM golang:1.8

# Install a more recent version of mercurial to avoid mismatching results
# between dep runs on a decently updated host system and the build container.
RUN awk '$1 ~ "^deb" { $3 = $3 "-backports"; print; exit }' /etc/apt/sources.list > /etc/apt/sources.list.d/backports.list && \
  DEBIAN_FRONTEND=noninteractive apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -t jessie-backports --yes --no-install-recommends mercurial=3.9.1-1~bpo8+1 && \
  rm -fr /var/lib/apt/lists/

RUN go get github.com/jteeuwen/go-bindata/... \
&& go get github.com/golang/lint/golint \
&& go get github.com/kisielk/errcheck \
&& go get github.com/client9/misspell/cmd/misspell \
&& go get github.com/golang/dep/cmd/dep

# Which docker version to test on
ARG DOCKER_VERSION=17.03.2

# Download docker
RUN mkdir -p /usr/local/bin \
    && curl -fL https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}-ce.tgz \
    | tar -xzC /usr/local/bin --transform 's#^.+/##x'

WORKDIR /go/src/github.com/containous/traefik
COPY . /go/src/github.com/containous/traefik
