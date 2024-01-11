# syntax=docker/dockerfile:1.2
FROM scratch

COPY script/ca-certificates.crt /etc/ssl/certs/

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/traefik /

EXPOSE 80
VOLUME ["/tmp"]

ENTRYPOINT ["/traefik"]
