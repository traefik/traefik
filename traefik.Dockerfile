ARG TRAEFIK_IMAGE_VERSION=dev

# ------------------------------------------------------------------------------

FROM traefik-backend:$TRAEFIK_IMAGE_VERSION as backend

# ------------------------------------------------------------------------------

## IMAGE
FROM alpine:3.10

RUN apk --no-cache --no-progress add bash curl ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=backend /go/src/github.com/containous/traefik/dist/traefik /

EXPOSE 80
VOLUME ["/tmp"]

ENTRYPOINT ["/traefik"]
