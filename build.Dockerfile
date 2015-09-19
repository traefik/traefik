FROM golang:1.5

RUN go get github.com/mitchellh/gox
RUN go get github.com/tcnksm/ghr
# Install dependencies
RUN go get github.com/BurntSushi/toml \
    && go get github.com/BurntSushi/ty/fun
RUN go get github.com/mailgun/oxy/forward \
    && go get github.com/mailgun/oxy/roundrobin \
    && go get github.com/mailgun/oxy/cbreaker \
    && go get github.com/mailgun/log \
    && go get github.com/mailgun/predicate
RUN go get github.com/gorilla/handlers \
    && go get github.com/gorilla/mux
RUN go get github.com/cenkalti/backoff \
    && go get github.com/codegangsta/negroni \
    && go get github.com/op/go-logging \
    && go get github.com/elazarl/go-bindata-assetfs \
    && go get github.com/leekchan/gtf \
    && go get github.com/thoas/stats \
    && go get github.com/tylerb/graceful \
    && go get github.com/unrolled/render
RUN go get github.com/fsouza/go-dockerclient \
    && go get github.com/gambol99/go-marathon
RUN go get gopkg.in/fsnotify.v1 \
    && go get gopkg.in/alecthomas/kingpin.v2

WORKDIR /go/src/github.com/emilevauge/traefik

COPY . /go/src/github.com/emilevauge/traefik
