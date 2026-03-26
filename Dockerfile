# syntax=docker/dockerfile:1.2
FROM alpine:3.23


ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/traefik /

RUN apk add --no-cache --no-progress ca-certificates libcap-setcap tzdata && \
  setcap    cap_net_bind_service=+ep /traefik && \
  setcap -v cap_net_bind_service=+ep /traefik

EXPOSE 80
VOLUME ["/tmp"]

ENTRYPOINT ["/traefik"]
