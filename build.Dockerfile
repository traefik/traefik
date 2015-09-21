FROM golang:1.5

RUN go get github.com/tools/godep
RUN go get github.com/mitchellh/gox
RUN go get github.com/tcnksm/ghr

ENV PATH /go/src/github.com/emilevauge/traefik/Godeps/_workspace/bin:$PATH

WORKDIR /go/src/github.com/emilevauge/traefik

RUN mkdir Godeps
COPY Godeps/Godeps.json Godeps/
RUN godep restore

COPY . /go/src/github.com/emilevauge/traefik
