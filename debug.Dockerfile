FROM ubuntu:latest
# Feel free to add below any helpful dependency for debugging
RUN apt-get update && apt-get install -y --no-install-recommends lsof iproute2
COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik /
EXPOSE 80
VOLUME ["/tmp"]
ENTRYPOINT ["/traefik"]
