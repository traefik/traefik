FROM scratch
COPY dist/traefik_linux-386 /traefik
ENTRYPOINT ["/traefik"]
