FROM alpine:3.14
# Feel free to add below any helpful dependency for debugging.
# iproute2 is for ss.
RUN apk --no-cache --no-progress add bash curl ca-certificates tzdata lsof iproute2 \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*
COPY dist/traefik /
EXPOSE 80
VOLUME ["/tmp"]
ENTRYPOINT ["/traefik"]
