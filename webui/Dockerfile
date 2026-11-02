# syntax=docker/dockerfile:1
FROM node:26-alpine3.24
# Current Active LTS release according to (https://nodejs.org/en/about/releases/)

RUN npm i -g yarn@1.22.22

ENV WEBUI_DIR=/src/webui
RUN mkdir -p $WEBUI_DIR

COPY package.json $WEBUI_DIR/
COPY yarn.lock $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN yarn install

COPY . $WEBUI_DIR/

EXPOSE 8080

RUN yarn lint
