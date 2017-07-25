FROM    golang:1.8-alpine

RUN     apk add -U git

RUN     go get -u gopkg.in/alecthomas/gometalinter.v1 && \
        mv /go/bin/gometalinter.v1 /usr/local/bin/gometalinter && \
        gometalinter --install

WORKDIR /go/src/github.com/docker/cli
ENTRYPOINT ["/usr/local/bin/gometalinter"]
CMD ["--config=gometalinter.json", "./..."]
