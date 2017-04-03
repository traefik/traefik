FROM golang:1.8

RUN go get github.com/jteeuwen/go-bindata/...

WORKDIR /go/src/github.com/containous/traefik
COPY . /go/src/github.com/containous/traefik
