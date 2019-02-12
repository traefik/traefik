FROM node:10-alpine

# Please note that the markdown spellchecker is installed from a fork
# until https://github.com/lukeapage/node-markdown-spellcheck/pull/114 merged
RUN apk --no-cache --no-progress add \
    ca-certificates \
    curl \
    findutils \
    git \
    ruby-bigdecimal \
    ruby-etc \
    ruby-ffi \
    ruby-json \
    ruby-nokogiri=1.8.3-r0 \
    tini \
  && gem install --no-document html-proofer -v 3.9.3 \
  && npm install write-good markdownlint markdownlint-cli --global

COPY ./scripts/verify.sh /verify.sh
COPY ./scripts/lint.sh /lint.sh

WORKDIR /app
VOLUME ["/tmp","/app"]

ENTRYPOINT ["/sbin/tini","-g","sh"]
