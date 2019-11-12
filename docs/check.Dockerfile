
FROM alpine:3.10 as alpine

RUN apk --no-cache --no-progress add \
    libcurl \
    ruby \
    ruby-bigdecimal \
    ruby-etc \
    ruby-ffi \
    ruby-json \
    ruby-nokogiri
RUN NOKOGIRI_USE_SYSTEM_LIBRARIES=true gem install --no-document html-proofer -v 3.13.0

# After Ruby, some NodeJS YAY!
RUN apk --no-cache --no-progress add \
    git \
    nodejs \
    npm \
  && npm install markdownlint@0.12.0 markdownlint-cli@0.13.0 --global

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
