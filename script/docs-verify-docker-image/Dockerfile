FROM alpine:3.8

RUN apk --no-cache --no-progress add \
    ca-certificates \
    curl \
    findutils \
    ruby-bigdecimal \
    ruby-etc \
    ruby-ffi \
    ruby-json \
    ruby-nokogiri=1.8.3-r0 \
    tini \
  && gem install --no-document html-proofer -v 3.9.3

COPY ./validate.sh /validate.sh

WORKDIR /app
VOLUME ["/tmp","/app"]

ENTRYPOINT ["/sbin/tini","-g","sh"]
CMD ["/validate.sh"]
