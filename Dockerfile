FROM scratch
COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik /
COPY traefik.sample.yml /etc/traefik/treaefik.toml
EXPOSE 80 8080
VOLUME ["/tmp"]
ENTRYPOINT ["/traefik"]
