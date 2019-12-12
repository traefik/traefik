FROM scratch
COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik /
EXPOSE 80
VOLUME ["/tmp"]
## Fix nssswitch not looking at host file (See https://stackoverflow.com/questions/49476452/traefik-forwarding-to-a-host-and-overriding-ip?answertab=oldest#tab-top)
RUN echo "hosts: files dns" > /etc/nsswitch.conf
ENTRYPOINT ["/traefik"]
