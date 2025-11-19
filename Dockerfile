# syntax=docker/dockerfile:1.2
FROM alpine:3.22

RUN apk add --no-cache --no-progress ca-certificates tzdata

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/traefik /

RUN addgroup -g 65532 -S traefik && \
    adduser -u 65532 -D -S -G traefik traefik

EXPOSE 80
VOLUME ["/tmp"]

USER traefik
ENTRYPOINT ["/traefik"]
