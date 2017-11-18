FROM alpine
COPY docker-entrypoint.sh /
COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik /
EXPOSE 80
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["traefik"]
