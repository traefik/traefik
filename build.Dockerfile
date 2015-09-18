FROM golang:1.5

RUN go get github.com/tools/godep
RUN go get github.com/mitchellh/gox
RUN go get github.com/tcnksm/ghr

ENV GOPATH /go/src/github.com/emilevauge/traefik/Godeps/_workspace:/go
ENV PATH /go/src/github.com/emilevauge/traefik/Godeps/_workspace/bin:$PATH

WORKDIR /go/src/github.com/emilevauge/traefik

COPY . /go/src/github.com/emilevauge/traefik
