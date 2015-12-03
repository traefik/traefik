FROM scratch
COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik /
EXPOSE 80
ENTRYPOINT ["/traefik"]
