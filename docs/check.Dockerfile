FROM alpine:3.14 as alpine

RUN apk --no-cache --no-progress add \
    build-base \
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

RUN gem install nokogiri --version 1.13.3 --no-document -- --use-system-libraries
RUN gem install html-proofer --version 3.19.3 --no-document -- --use-system-libraries

# After Ruby, some NodeJS YAY!
RUN apk --no-cache --no-progress add \
    git \
    nodejs \
    npm

# To handle 'not get uid/gid'
RUN npm config set unsafe-perm true

RUN npm install --global \
    markdownlint@0.22.0 \
    markdownlint-cli@0.26.0

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
