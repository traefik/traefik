
FROM    golang:1.8-alpine

RUN     apk add -U git make bash coreutils

RUN     go get github.com/LK4D4/vndr && \
        cp /go/bin/vndr /usr/bin && \
        rm -rf /go/src/* /go/pkg/* /go/bin/*

RUN     go get github.com/jteeuwen/go-bindata/go-bindata && \
        cp /go/bin/go-bindata /usr/bin && \
        rm -rf /go/src/* /go/pkg/* /go/bin/*

RUN     go get github.com/dnephin/filewatcher && \
        cp /go/bin/filewatcher /usr/bin/ && \
        rm -rf /go/src/* /go/pkg/* /go/bin/*

ENV     CGO_ENABLED=0
WORKDIR /go/src/github.com/docker/cli
CMD     sh
