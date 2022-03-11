# WEBUI
FROM node:12.11 as webui

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

COPY ./webui/ $WEBUI_DIR/

WORKDIR $WEBUI_DIR

RUN yarn install
RUN yarn build

# BUILD
FROM golang:1.17-alpine as gobuild

RUN apk --no-cache --no-progress add git mercurial bash gcc musl-dev curl tar ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /go/src/github.com/traefik/traefik

# Download go modules
COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on GOPROXY=https://proxy.golang.org go mod download

COPY . /go/src/github.com/traefik/traefik

RUN rm -rf /go/src/github.com/traefik/traefik/webui/static/
COPY --from=webui /src/webui/static/ /go/src/github.com/traefik/traefik/webui/static/

RUN ./script/make.sh generate binary

## IMAGE
FROM alpine:3.14

RUN apk --no-cache --no-progress add bash curl ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=gobuild /go/src/github.com/traefik/traefik/dist/traefik /

EXPOSE 80
VOLUME ["/tmp"]

ENTRYPOINT ["/traefik"]
