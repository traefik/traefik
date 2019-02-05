FROM node:8.15.0

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

COPY package.json $WEBUI_DIR/
COPY yarn.lock $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN yarn install

COPY . $WEBUI_DIR/

EXPOSE 8080

RUN yarn lint
