
FROM alpine:3.9 as alpine

# The "build-dependencies" virtual package provides build tools for html-proofer installation.
# It compile ruby-nokogiri, because alpine native version is always out of date
# This virtual package is cleaned at the end.
RUN apk --no-cache --no-progress add \
    libcurl \
    ruby \
    ruby-bigdecimal \
    ruby-etc \
    ruby-ffi \
    ruby-json \
  && apk add --no-cache --virtual build-dependencies \
    build-base \
    libcurl \
    libxml2-dev \
    libxslt-dev \
    ruby-dev \
  && gem install --no-document html-proofer -v 3.10.2 \
  && apk del build-dependencies

# After Ruby, some NodeJS YAY!
RUN apk --no-cache --no-progress add \
    git \
    nodejs \
    npm \
  && npm install write-good@1.0.1 markdownlint@0.12.0 markdownlint-cli@0.13.0 --global

# Finally the shell tools we need for later
# tini helps to terminate properly all the parallelized tasks when sending CTRL-C
RUN apk --no-cache --no-progress add \
    ca-certificates \
    curl \
    tini

COPY ./scripts/verify.sh /verify.sh
COPY ./scripts/lint.sh /lint.sh

WORKDIR /app
VOLUME ["/tmp","/app"]

ENTRYPOINT ["/sbin/tini","-g","sh"]
