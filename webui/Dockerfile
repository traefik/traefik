FROM node:8.15.0

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

COPY package.json $WEBUI_DIR/
COPY package-lock.json $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm install

COPY . $WEBUI_DIR/

EXPOSE 8080

RUN npm run lint
