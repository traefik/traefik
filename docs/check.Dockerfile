FROM alpine:3.18 as alpine

RUN apk --no-cache --no-progress add \
    build-base \
    gcompat \
    libcurl \
    libxml2-dev \
    libxslt-dev \
    ruby \
    ruby-bigdecimal \
    ruby-dev \
    ruby-etc \
    ruby-ffi \
    ruby-json \
    zlib-dev

RUN gem install nokogiri --version 1.15.3 --no-document -- --use-system-libraries
RUN gem install html-proofer --version 5.0.7 --no-document -- --use-system-libraries

# After Ruby, some NodeJS YAY!
RUN apk --no-cache --no-progress add \
    git \
    nodejs \
    npm

RUN npm install --global \
    markdownlint@0.29.0 \
    markdownlint-cli@0.35.0

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
