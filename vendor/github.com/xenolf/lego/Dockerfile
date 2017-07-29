FROM alpine:3.4

ENV GOPATH /go

RUN apk update && apk add ca-certificates go git && \
    rm -rf /var/cache/apk/* && \
    go get -u github.com/xenolf/lego && \
    cd /go/src/github.com/xenolf/lego && \
    go build -o /usr/bin/lego . && \
    apk del go git && \
    rm -rf /var/cache/apk/* && \
    rm -rf /go

ENTRYPOINT [ "/usr/bin/lego" ]
