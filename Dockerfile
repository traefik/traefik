FROM scratch
COPY dist/traefik /
ENTRYPOINT ["/traefik"]
