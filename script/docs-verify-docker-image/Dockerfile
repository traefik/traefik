FROM alpine:3.7

RUN apk --no-cache --no-progress add \
    ca-certificates \
    curl \
    findutils \
    ruby-bigdecimal \
    ruby-ffi \
    ruby-json \
    ruby-nokogiri \
    tini \
  && gem install --no-document html-proofer

COPY ./validate.sh /validate.sh

WORKDIR /app
VOLUME ["/tmp","/app"]

ENTRYPOINT ["/sbin/tini","-g","sh"]
CMD ["/validate.sh"]
