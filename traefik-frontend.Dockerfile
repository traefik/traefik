FROM node:12.11

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

COPY webui/package.json $WEBUI_DIR/
COPY webui/package-lock.json $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm install

COPY webui/ $WEBUI_DIR/

RUN npm run lint
RUN npm run build:nc